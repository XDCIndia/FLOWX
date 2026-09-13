package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fluxa/fluxa/internal/compliance"
	"github.com/fluxa/fluxa/internal/config"
	"github.com/fluxa/fluxa/internal/fees"
	"github.com/fluxa/fluxa/internal/chain/xdc"
	"github.com/fluxa/fluxa/internal/postgres"
	"github.com/fluxa/fluxa/internal/queue"
	"github.com/fluxa/fluxa/internal/schedule"
	"github.com/fluxa/fluxa/internal/settlement"
	"github.com/fluxa/fluxa/internal/transfer"
	"github.com/fluxa/fluxa/internal/webhook"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	if !cfg.WorkerEnabled {
		log.Info().Msg("worker disabled for this region")
		return
	}

	if cfg.Env == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("connect to database")
	}
	defer db.Close()
	var replica *pgxpool.Pool
	if cfg.ReplicaDatabaseURL != "" {
		replica, err = postgres.New(ctx, cfg.ReplicaDatabaseURL)
		if err != nil {
			log.Warn().Err(err).Msg("connect to read replica; reads will use primary")
		}
		if replica != nil {
			defer replica.Close()
		}
	}
	repoDB := postgres.NewReplicaAwareDB(db, replica)

	walletRepo := postgres.NewWalletRepo(repoDB)
	txRepo := postgres.NewTransactionRepo(repoDB)
	feeRepo := postgres.NewFeeRepo(repoDB)
	webhookRepo := postgres.NewWebhookRepo(repoDB)
	scheduleRepo := postgres.NewScheduleRepo(repoDB)
	complianceRepo := postgres.NewComplianceRepo(repoDB).WithPrimary(db)

	feeSvc := fees.NewService(feeRepo)

	// Settlement engine: XDC Apothem backend.
	xdcClient, err := xdc.New(ctx, cfg.XDCRPCURL, cfg.XDCChainID)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialise XDC chain client")
	}
	log.Info().Str("rpc", cfg.XDCRPCURL).Int64("chain_id", cfg.XDCChainID).
		Msg("settlement backend: XDC (Apothem testnet model)")
	submitter := settlement.NewXDCEngine(txRepo, walletRepo, feeSvc, xdcClient, cfg.MasterEncryptionKey, cfg.PlatformFeeWalletPublicKey)
	settlementWorker := settlement.NewWorker(submitter)

	asynqOpt, err := queue.AsynqRedisOptions(cfg.RedisURL, cfg.RedisSentinelMasterName, cfg.RedisSentinelAddrs, cfg.RedisSentinelPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("configure asynq redis")
	}
	qClient := queue.NewClientWithOptions(asynqOpt)
	defer qClient.Close()
	redisOpt, err := queue.RedisOptions(cfg.RedisURL, cfg.RedisSentinelMasterName, cfg.RedisSentinelAddrs, cfg.RedisSentinelPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("configure redis")
	}
	redisClient := redis.NewUniversalClient(redisOpt)
	defer redisClient.Close()
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			if err := redisClient.Set(ctx, "flowx:worker:heartbeat", time.Now().UTC().Format(time.RFC3339Nano), 30*time.Second).Err(); err != nil {
				log.Warn().Err(err).Msg("worker heartbeat failed")
			}
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}
	}()

	webhookSvc := webhook.NewService(webhookRepo, qClient)
	webhookWorker := webhook.NewWorker(webhookSvc)

	transferSvc := transfer.NewService(txRepo, walletRepo, feeSvc, qClient)

	// The worker screens too: scheduled payouts run here and go through
	// transfer.initiate() exactly like an API-initiated transfer, so leaving
	// the screener off would let them bypass compliance entirely.
	var complianceWorker *compliance.Worker
	if cfg.ComplianceEnabled {
		sanctionsSet := compliance.NewSanctionsSet()
		if err := sanctionsSet.LoadFromRepository(ctx, complianceRepo); err != nil {
			log.Error().Err(err).Msg("compliance: initial sanctions load failed; transfers will be held until it succeeds")
		}
		sanctionsSet.StartReloader(ctx, complianceRepo,
			time.Duration(cfg.ComplianceReloadMinutes)*time.Minute)

		structuringUnit, err := decimal.NewFromString(cfg.ComplianceStructuringUnit)
		if err != nil {
			log.Fatal().Err(err).Msg("parse COMPLIANCE_STRUCTURING_UNIT")
		}

		screener := compliance.NewCompositeScreener(
			compliance.NewSanctionsScreener(sanctionsSet, cfg.ComplianceFuzzyThreshold),
			compliance.NewVelocityScreener(complianceRepo, compliance.VelocityConfig{
				Window:           time.Duration(cfg.ComplianceVelocityWindowMin) * time.Minute,
				MaxTransfers:     cfg.ComplianceVelocityMax,
				StructuringUnit:  structuringUnit,
				RoundTripWindow:  time.Duration(cfg.ComplianceRoundTripMin) * time.Minute,
				PlatformWalletID: cfg.PlatformWalletID,
			}),
		)

		complianceSvc := compliance.NewService(complianceRepo, screener, sanctionsSet, txRepo, qClient, webhookSvc)
		transferSvc = transferSvc.WithScreener(complianceSvc)
		complianceWorker = compliance.NewWorker(
			complianceRepo,
			compliance.NewHTTPSDNSource(cfg.OFACSDNURL, nil),
			sanctionsSet,
			webhookSvc,
		)
	}

	scheduleWorker := schedule.NewWorker(scheduleRepo, transferSvc)

	srv := asynq.NewServer(asynqOpt, asynq.Config{

		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeProcessTransfer, settlementWorker.HandleProcessTransfer)
	mux.HandleFunc(queue.TypeWebhookDeliver, webhookWorker.HandleDeliver)
	mux.HandleFunc(queue.TypeRunSchedules, scheduleWorker.HandleRunSchedules)
	if complianceWorker != nil {
		mux.HandleFunc(queue.TypeRefreshSanctions, complianceWorker.HandleRefreshSanctions)
	}

	scheduler := asynq.NewScheduler(asynqOpt, nil)

	// Scheduled payouts are checked every minute — matches the acceptance
	// window (fires within ±1 minute of next_run_at) without needing a
	// dedicated ticker.
	scheduleTask := asynq.NewTask(queue.TypeRunSchedules, nil)
	if _, err := scheduler.Register("@every 1m", scheduleTask); err != nil {
		log.Fatal().Err(err).Msg("register schedule run scheduler")
	}

	// The OFAC SDN list is republished on business days; a daily refresh on the
	// low queue keeps every process's in-memory set current via
	// sanctions_entities without competing with live settlement.
	if complianceWorker != nil {
		sanctionsTask := asynq.NewTask(queue.TypeRefreshSanctions, nil, asynq.Queue("low"))
		if _, err := scheduler.Register("@daily", sanctionsTask); err != nil {
			log.Fatal().Err(err).Msg("register sanctions refresh scheduler")
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := scheduler.Run(); err != nil {
			log.Error().Err(err).Msg("scheduler error")
		}
	}()

	go func() {
		log.Info().Msg("flowx worker starting")
		if err := srv.Run(mux); err != nil {
			log.Error().Err(err).Msg("worker stopped")
		}
	}()

	<-quit
	log.Info().Msg("flowx worker shutting down")
	cancel()
	srv.Shutdown()
	scheduler.Shutdown()
}
