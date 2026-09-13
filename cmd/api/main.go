package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fluxa/fluxa/internal/apikey"
	"github.com/fluxa/fluxa/internal/auth"
	"github.com/fluxa/fluxa/internal/batch"
	"github.com/fluxa/fluxa/internal/chain/xdc"
	"github.com/fluxa/fluxa/internal/compliance"
	"github.com/fluxa/fluxa/internal/config"
	"github.com/fluxa/fluxa/internal/fees"
	"github.com/fluxa/fluxa/internal/fiat"
	"github.com/fluxa/fluxa/internal/fiat/flutterwave"
	striperail "github.com/fluxa/fluxa/internal/fiat/stripe"
	"github.com/fluxa/fluxa/internal/fx"
	"github.com/fluxa/fluxa/internal/org"
	"github.com/fluxa/fluxa/internal/postgres"
	"github.com/fluxa/fluxa/internal/queue"

	"github.com/fluxa/fluxa/internal/routing"
	"github.com/fluxa/fluxa/internal/schedule"
	"github.com/fluxa/fluxa/internal/server"
	"github.com/fluxa/fluxa/internal/server/idempotency"
	"github.com/fluxa/fluxa/internal/settlement"
	"github.com/fluxa/fluxa/internal/chain/xdc"
	"github.com/fluxa/fluxa/internal/transfer"
	"github.com/fluxa/fluxa/internal/wallet"
	"github.com/fluxa/fluxa/internal/webhook"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

func main() {
	migrateOnly := flag.Bool("migrate-only", false, "run migrations and exit")
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	if cfg.Env == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := postgres.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		log.Fatal().Err(err).Msg("run migrations")
	}
	if *migrateOnly {
		log.Info().Msg("migrations complete")
		return
	}

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

	redisOpt, err := queue.RedisOptions(cfg.RedisURL, cfg.RedisSentinelMasterName, cfg.RedisSentinelAddrs, cfg.RedisSentinelPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("parse redis url")
	}
	redisClient := redis.NewUniversalClient(redisOpt)
	defer redisClient.Close()

	tenantRepo := postgres.NewTenantRepo(repoDB)
	userRepo := postgres.NewUserRepo(repoDB)
	orgRepo := postgres.NewOrgRepo(repoDB)

	walletRepo := postgres.NewWalletRepo(repoDB)
	txRepo := postgres.NewTransactionRepo(repoDB)

	convRepo := postgres.NewConversionRepo(repoDB)
	feeRepo := postgres.NewFeeRepo(repoDB)
	apiKeyRepo := postgres.NewAPIKeyRepo(repoDB)
	fiatRepo := postgres.NewFiatRepo(repoDB)
	webhookRepo := postgres.NewWebhookRepo(repoDB)
	fxQuoteRepo := postgres.NewFXQuoteRepo(repoDB)
	batchRepo := postgres.NewBatchRepo(repoDB)
	scheduleRepo := postgres.NewScheduleRepo(repoDB)
	idempotencyRepo := postgres.NewIdempotencyRepo(repoDB)
	complianceRepo := postgres.NewComplianceRepo(repoDB).WithPrimary(db)
	idemMW := idempotency.Middleware(idempotencyRepo)

	asynqOpt, err := queue.AsynqRedisOptions(cfg.RedisURL, cfg.RedisSentinelMasterName, cfg.RedisSentinelAddrs, cfg.RedisSentinelPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("configure asynq redis")
	}
	queueClient := queue.NewClientWithOptions(asynqOpt)
	defer queueClient.Close()

	jwtSecretBytes := []byte(cfg.JWTSecret)

	authSvc := auth.NewService(userRepo, tenantRepo, orgRepo, jwtSecretBytes)
	orgSvc := org.NewService(orgRepo, userRepo, tenantRepo, jwtSecretBytes)

	feeSvc := fees.NewService(feeRepo)

	// Wallet service: XDC (Apothem testnet model) backend.
	xdcClient, err := xdc.New(context.Background(), cfg.XDCRPCURL, cfg.XDCChainID)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialise XDC chain client")
	}
	log.Info().Str("rpc", cfg.XDCRPCURL).Int64("chain_id", cfg.XDCChainID).
		Msg("wallet backend: XDC (Apothem testnet model)")
	walletSvc := wallet.NewXDCService(walletRepo, txRepo, xdcClient, cfg.MasterEncryptionKey, cfg.XDCTreasurySecretKey, tenantRepo)

	transferSvc := transfer.NewService(txRepo, walletRepo, feeSvc, queueClient, tenantRepo)
	webhookSvc := webhook.NewService(webhookRepo, queueClient, tenantRepo)

	// Compliance screening sits in front of settlement, so it is wired before
	// the services that initiate transfers. When disabled, no screener is
	// attached and transfers keep their pre-compliance behaviour.
	var complianceHandler *compliance.Handler
	if cfg.ComplianceEnabled {
		sanctionsSet := compliance.NewSanctionsSet()

		// Not fatal: screening fails closed, so an API that boots before the
		// first SDN refresh holds transfers for review rather than clearing
		// them. Log loudly and carry on.
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

		complianceSvc := compliance.NewService(complianceRepo, screener, sanctionsSet, txRepo, queueClient, webhookSvc)
		complianceHandler = compliance.NewHandler(complianceSvc)
		transferSvc = transferSvc.WithScreener(complianceSvc)
	}

	batchSvc := batch.NewService(batchRepo, txRepo, transferSvc)
	scheduleSvc := schedule.NewService(scheduleRepo, walletRepo)

	// FX providers: CoinGecko supplies the USDC<->XDC spot rate.
	fxProviders := []fx.Provider{fx.NewCoinGeckoProvider("")}
	// Fixed fiat→crypto rates for the testnet model (FIAT_STATIC_RATES),
	// e.g. NGN-USDC / NGN-TXDC. Testnet scaffolding, not for production.
	if cfg.FIATStaticRates != "" {
		fxProviders = append(fxProviders, fx.NewStaticProvider(fx.ParseStaticRates(cfg.FIATStaticRates)))
	}
	fxSvc := fx.NewService(
		walletRepo, convRepo, fxQuoteRepo,
		feeSvc, redisClient,
		fxProviders, cfg.FXSpreadBps,
	)
	walletSvc.WithFXService(fxSvc)
	fx.SetXDC(fxSvc, xdcClient, cfg.XDCTreasurySecretKey)

	// fiat.Service drives exactly one rail, selected by FIAT_RAIL
	// ("flutterwave" default, "stripe" optional). The Yellow Card provider
	// (internal/fiat/yellowcard) is implemented but not wired; per-request
	// provider selection is future work.
	//
	// The credit asset is TXDC: the XDC settlement engine settles the native
	// asset in the testnet model.
	fiatCreditAsset := "TXDC"
	fiatProviderName := cfg.FIATRail
	if fiatProviderName == "" {
		fiatProviderName = "flutterwave"
	}
	var fiatRail fiat.Rail
	switch fiatProviderName {
	case "stripe":
		fiatRail = striperail.NewRail(cfg.StripeSecretKey, cfg.StripeWebhookSecret, cfg.StripeSuccessURL)
	case "flutterwave":
		fwProvider := flutterwave.NewProvider(cfg.FlutterwaveSecretKey, cfg.FlutterwaveWebhookHash, cfg.AppBaseURL)
		fiatRail = fiat.NewRailAdapter(fwProvider)
	default:
		log.Fatal().Str("fiat_rail", fiatProviderName).Msg("unknown FIAT_RAIL (want flutterwave or stripe)")
	}
	log.Info().Str("fiat_rail", fiatProviderName).Str("credit_asset", fiatCreditAsset).Msg("fiat rail configured")

	fiatSvc := fiat.NewService(fiatRepo, fiatRail, fxSvc, transferSvc, cfg.PlatformWalletID, fiatProviderName, fiatCreditAsset)

	// Embedded worker (WORKER_ENABLED=true): settlement engine.
	submitter := settlement.NewXDCEngine(txRepo, walletRepo, feeSvc, xdcClient, cfg.MasterEncryptionKey, cfg.PlatformFeeWalletPublicKey)
	settlementWorker := settlement.NewWorker(submitter)

	asynqSrv := asynq.NewServer(asynqOpt, asynq.Config{
		Concurrency: 5,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
	})
	asynqMux := asynq.NewServeMux()
	asynqMux.HandleFunc(queue.TypeProcessTransfer, settlementWorker.HandleProcessTransfer)

	if cfg.WorkerEnabled {
		go func() {
			log.Info().Msg("flowx api: settlement asynq consumer starting")
			if err := asynqSrv.Run(asynqMux); err != nil {
				log.Error().Err(err).Msg("flowx api: asynq consumer stopped")
			}
		}()
	}

	authHandler := auth.NewHandler(authSvc)
	orgHandler := org.NewHandler(orgSvc)
	walletHandler := wallet.NewHandler(walletSvc).
		WithIdempotency(idemMW).
		WithFaucet(cfg.TestnetFaucetEnabled, cfg.FaucetMaxAmount)
	transferHandler := transfer.NewHandler(transferSvc).WithIdempotency(idemMW)
	fxHandler := fx.NewHandler(fxSvc).WithIdempotency(idemMW)
	routingHandler := routing.NewHandler(cfg.StripeSecretKey)
	// Create XDC client for blockchain route
	var xdcRouteClient *xdc.Client
	xdcRouteClient, err = xdc.New(context.Background(), cfg.XDCRPCURL, cfg.XDCChainID)
	if err != nil {
		log.Warn().Err(err).Msg("xdc bridge route: failed to init chain client")
	}
	// Treasury wallet sends TXDC; second wallet is demo recipient
	recipientAddr := "0x7c42b69b8668504cbbcd3dc9506f0493f4f75fed"
	routingHandler.RegisterRoute(routing.NewXDCBridgeRoute(xdcRouteClient, cfg.XDCTreasurySecretKey, recipientAddr, fxSvc))
	routingHandler.RegisterRoute(fiat.NewPaymentNetworkRoute("INR-EUR", fxSvc))
	routingHandler.RegisterRoute(fiat.NewPaymentNetworkRoute("EUR-INR", fxSvc))
	routingHandler.RegisterRoute(fiat.NewPaymentNetworkRoute("INR-USDC", fxSvc))
	routingHandler.RegisterRoute(fiat.NewPaymentNetworkRoute("USDC-INR", fxSvc))
	routingHandler.RegisterRoute(fiat.NewPaymentNetworkRoute("INR-TXDC", fxSvc))
	routingHandler.RegisterRoute(fiat.NewPaymentNetworkRoute("TXDC-INR", fxSvc))
	routingHandler.RegisterRoute(routing.NewStripeBankRoute(cfg.StripeSecretKey, fxSvc))
	routingHandler.RegisterRoute(routing.NewOnChainXDCRoute())
	fiatHandler := fiat.NewHandler(fiatSvc)
	feeHandler := fees.NewHandler(feeSvc)
	apikeyHandler := apikey.NewHandler(apiKeyRepo)
	webhookHandler := webhook.NewHandler(webhookSvc)
	batchHandler := batch.NewHandler(batchSvc).WithIdempotency(idemMW)
	scheduleHandler := schedule.NewHandler(scheduleSvc)

	srv := server.New(
		authHandler, orgHandler, walletHandler, transferHandler, fxHandler, fiatHandler,
		feeHandler, apikeyHandler, apiKeyRepo, txRepo,
		webhookHandler, batchHandler, scheduleHandler, complianceHandler, routingHandler, cfg.CORSOrigins, jwtSecretBytes, cfg.Port,
		map[string]server.DependencyCheck{
			"postgres": db.Ping,
			"replica":  func(ctx context.Context) error { return repoDB.ReplicaAvailable(ctx) },
			"redis":    func(ctx context.Context) error { return redisClient.Ping(ctx).Err() },

			"chain": func(ctx context.Context) error {
				if cfg.XDCRPCURL != "" {
					req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.XDCRPCURL, nil)
					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						return err
					}
					resp.Body.Close()
					return nil
				}
				return nil
			},
			"worker": func(ctx context.Context) error { return nil },
		},

		orgRepo,
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("port", cfg.Port).Msg("fluxa api starting")
		if err := srv.Start(); err != nil {
			log.Error().Err(err).Msg("server stopped")
		}
	}()

	<-quit
	log.Info().Msg("shutting down")

	cancel()

	asynqSrv.Shutdown()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}

	log.Info().Msg("goodbye")
}
