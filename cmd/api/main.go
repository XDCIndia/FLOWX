package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fluxa/fluxa/internal/alerting"
	"github.com/fluxa/fluxa/internal/anchor"
	"github.com/fluxa/fluxa/internal/apikey"
	"github.com/fluxa/fluxa/internal/assets"
	"github.com/fluxa/fluxa/internal/auth"
	"github.com/fluxa/fluxa/internal/batch"
	"github.com/fluxa/fluxa/internal/compliance"
	"github.com/fluxa/fluxa/internal/config"
	"github.com/fluxa/fluxa/internal/domain"
	"github.com/fluxa/fluxa/internal/fees"
	"github.com/fluxa/fluxa/internal/fiat"
	"github.com/fluxa/fluxa/internal/routing"
	"github.com/fluxa/fluxa/internal/fiat/flutterwave"
	striperail "github.com/fluxa/fluxa/internal/fiat/stripe"
	"github.com/fluxa/fluxa/internal/fx"
	"github.com/fluxa/fluxa/internal/indexer"
	"github.com/fluxa/fluxa/internal/org"
	"github.com/fluxa/fluxa/internal/postgres"
	"github.com/fluxa/fluxa/internal/queue"
	"github.com/fluxa/fluxa/internal/reconcile"
	"github.com/fluxa/fluxa/internal/schedule"
	"github.com/fluxa/fluxa/internal/server"
	"github.com/fluxa/fluxa/internal/server/idempotency"
	"github.com/fluxa/fluxa/internal/settlement"
	"github.com/fluxa/fluxa/internal/chain/xdc"
	"github.com/fluxa/fluxa/internal/stellar"
	"github.com/fluxa/fluxa/internal/transfer"
	"github.com/fluxa/fluxa/internal/treasury"
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
	reconcileRepo := postgres.NewReconcileRepo(repoDB)
	fxQuoteRepo := postgres.NewFXQuoteRepo(repoDB)
	batchRepo := postgres.NewBatchRepo(repoDB)
	scheduleRepo := postgres.NewScheduleRepo(repoDB)
	anchorRepo := postgres.NewAnchorRepo(repoDB)
	treasuryRepo := postgres.NewTreasuryRepo(repoDB)
	idempotencyRepo := postgres.NewIdempotencyRepo(repoDB)
	complianceRepo := postgres.NewComplianceRepo(repoDB).WithPrimary(db)
	idemMW := idempotency.Middleware(idempotencyRepo)

	stellarClient := stellar.NewClient(cfg.StellarHorizonURL, cfg.StellarNetwork)
	signer := stellar.NewEnvSigner(cfg.MasterEncryptionKey, cfg.StellarNetwork)

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
	// Wallet service: Stellar backend (default) or XDC Apothem backend
	// selected by CHAIN_BACKEND. See docs/xdc-migration-plan.md.
	var walletSvc wallet.Service
	if cfg.ChainBackend == "xdc" {
		xdcClient, err := xdc.New(context.Background(), cfg.XDCRPCURL, cfg.XDCChainID)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to initialise XDC chain client")
		}
		log.Info().Str("rpc", cfg.XDCRPCURL).Int64("chain_id", cfg.XDCChainID).
			Msg("wallet backend: XDC (Apothem testnet model)")
		walletSvc = wallet.NewXDCService(walletRepo, txRepo, xdcClient, cfg.MasterEncryptionKey, cfg.XDCTreasurySecretKey, tenantRepo)
	} else {
		walletSvc = wallet.NewService(walletRepo, stellarClient, cfg.MasterEncryptionKey, tenantRepo).
			WithSigner(signer).
			WithIssuers(cfg.StellarUSDCIssuer, cfg.StellarEURCIssuer)
	}
	transferSvc := transfer.NewService(txRepo, walletRepo, feeSvc, queueClient, tenantRepo)
	if cfg.ChainBackend != "xdc" {
		transferSvc = transferSvc.WithStellarClient(stellarClient)
	}
	webhookSvc := webhook.NewService(webhookRepo, queueClient, tenantRepo)

	// Compliance screening sits in front of settlement, so it is wired before
	// the services that initiate transfers. When disabled, no screener is
	// attached and transfers keep their pre-compliance behaviour.
	var complianceHandler *compliance.Handler
	if cfg.ComplianceEnabled {
		sanctionsSet := compliance.NewSanctionsSet()

		// Not fatal, unlike anchorRegistry.Load: screening fails closed, so an
		// API that boots before the first SDN refresh holds transfers for
		// review rather than clearing them. Log loudly and carry on.
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

	issuers := map[string]string{
		"USDC": cfg.StellarUSDCIssuer,
		"EURC": cfg.StellarEURCIssuer,
	}
	// USDC/EURC pairs cover the stablecoin corridor; XLM pairs give the
	// FX page a pair with real liquidity on both Stellar testnet and the
	// XDC backend's rate feed (the order book is Stellar DEX data).
	horizonProvider := fx.NewHorizonProvider(cfg.StellarHorizonURL, []string{"USDC-EURC", "EURC-USDC", "USDC-XLM", "XLM-USDC"}, issuers)
	fxProviders := []fx.Provider{horizonProvider}
	if cfg.ChainBackend == "xdc" {
		// XDC is not a Stellar asset, so the Horizon order book cannot
		// price it. CoinGecko supplies the USDC<->XDC spot rate instead.
		fxProviders = append(fxProviders, fx.NewCoinGeckoProvider(""))
	}
	// Fixed fiat→crypto rates for the testnet model (FIAT_STATIC_RATES),
	// e.g. NGN-USDC / NGN-TXDC. Testnet scaffolding, not for production.
	if cfg.FIATStaticRates != "" {
		fxProviders = append(fxProviders, fx.NewStaticProvider(fx.ParseStaticRates(cfg.FIATStaticRates)))
	}
	fxSvc := fx.NewService(
		walletRepo, convRepo, fxQuoteRepo,
		feeSvc, stellarClient, redisClient,
		cfg.StellarUSDCIssuer, fxProviders, cfg.FXSpreadBps,
	)
	walletSvc.WithFXService(fxSvc)

	// fiat.Service drives exactly one rail, selected by FIAT_RAIL
	// ("flutterwave" default, "stripe" optional). The Yellow Card provider
	// (internal/fiat/yellowcard) is implemented but not wired; per-request
	// provider selection is future work.
	//
	// The credit asset follows the chain backend: USDC on Stellar, TXDC on
	// XDC (the XDC settlement engine settles the native asset in the
	// testnet model).
	fiatCreditAsset := "USDC"
	if cfg.ChainBackend == "xdc" {
		fiatCreditAsset = "TXDC"
	}
	var fiatRail fiat.Rail
	fiatProviderName := cfg.FIATRail
	if fiatProviderName == "" {
		fiatProviderName = "flutterwave"
	}
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

	anchorRegistry := anchor.NewRegistry(anchorRepo, nil)
	if err := anchorRegistry.Load(ctx); err != nil {
		log.Fatal().Err(err).Msg("load anchor registry")
	}
	anchorFiatSvc := fiat.NewAnchorFiatService(anchorRegistry, anchorRepo, walletRepo, cfg.MasterEncryptionKey, cfg.StellarNetwork)

	treasurySvc := treasury.NewService(
		treasuryRepo, stellarClient, fxSvc, webhookSvc,
		cfg.PlatformFeeWalletPublicKey, cfg.StellarNetwork, cfg.TreasurySecretKey,
		cfg.StellarUSDCIssuer, cfg.StellarEURCIssuer,
	)

	// Embedded worker (WORKER_ENABLED=true): settlement engine + indexer.
	// Same CHAIN_BACKEND branch as cmd/worker — Horizon-backed pieces are
	// Stellar-only and skipped on XDC.
	var submitter settlement.TransferSubmitter
	if cfg.ChainBackend == "xdc" {
		xdcClient, err := xdc.New(context.Background(), cfg.XDCRPCURL, cfg.XDCChainID)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to initialise XDC chain client")
		}
		submitter = settlement.NewXDCEngine(txRepo, walletRepo, feeSvc, xdcClient, cfg.MasterEncryptionKey, cfg.PlatformFeeWalletPublicKey)
	} else {
		submitter = settlement.NewEngine(
			txRepo, walletRepo, feeSvc, stellarClient, signer,
			cfg.StellarNetwork, map[string]string{
				"USDC": cfg.StellarUSDCIssuer,
				"EURC": cfg.StellarEURCIssuer,
			}, cfg.PlatformFeeWalletPublicKey,
		)
	}
	settlementWorker := settlement.NewWorker(submitter)

	var indexerWorker *indexer.Worker
	if cfg.ChainBackend != "xdc" {
		idx := indexer.New(walletRepo, txRepo, stellarClient)
		indexerWorker = indexer.NewWorker(idx)

		// Live Horizon SSE stream keeps local state in sync in near real time.
		// processPayment is idempotent (guarded by ExistsByTxHash), so running
		// this alongside cmd/worker's own stream is safe, just extra capacity.
		go func() {
			if err := idx.StreamAll(ctx, 1000, 0); err != nil {
				log.Error().Err(err).Msg("indexer: stream all wallets failed")
			}
		}()
	}

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
	if indexerWorker != nil {
		asynqMux.HandleFunc(queue.TypeSyncLedger, indexerWorker.HandleSyncLedger)
	}

	if cfg.WorkerEnabled {
		go func() {
			log.Info().Msg("flowx api: settlement/indexer asynq consumer starting")
			if err := asynqSrv.Run(asynqMux); err != nil {
				log.Error().Err(err).Msg("flowx api: asynq consumer stopped")
			}
		}()
	}

	alertClient := alerting.NewClient(cfg.AlertWebhookURL, "fluxa-api")
	reconcileSvc := reconcile.NewService(
		txRepo,
		reconcileRepo,
		walletRepo,
		stellarClient,
		alertClient,
		queueClient,
		webhookSvc,
		"fluxa-api",
		decimal.Zero,
		assets.NewRegistry(cfg.StellarUSDCIssuer, cfg.StellarEURCIssuer),
		cfg.PlatformFeeWalletPublicKey,
	)
	reconcileHandler := reconcile.NewHandler(reconcileSvc)

	authHandler := auth.NewHandler(authSvc)
	orgHandler := org.NewHandler(orgSvc)
	walletHandler := wallet.NewHandler(walletSvc).WithIdempotency(idemMW)

	// Contract wallets are opt-in: without an installed WASM hash the API keeps
	// serving custodial wallets only and the contract routes stay unregistered.
	if cfg.ContractWalletWasmHash != "" {
		sorobanClient := stellar.NewSorobanClient(cfg.SorobanRPCURL, cfg.StellarNetwork)
		spendingLimit, err := decimal.NewFromString(cfg.ContractWalletSpendingLimit)
		if err != nil {
			log.Fatal().Err(err).Msg("parse CONTRACT_WALLET_SPENDING_LIMIT")
		}
		contractSvc := wallet.NewContractWalletAdapter(
			walletRepo,
			sorobanClient,
			wallet.NewSorobanDeployer(sorobanClient, signer, cfg.ContractWalletWasmHash),
			wallet.NewSACResolver(cfg.StellarNetwork, cfg.StellarUSDCIssuer, cfg.StellarEURCIssuer),
			wallet.ContractWalletParams{
				RecoveryThreshold:     uint32(cfg.ContractWalletRecoveryQuota),
				SpendingLimit:         spendingLimit,
				SpendingWindowSeconds: uint64(cfg.ContractWalletWindowSeconds),
			},
		).WithTenantRepo(tenantRepo)
		contractSvc.WithSigner(signer)
		walletHandler = walletHandler.WithContractService(contractSvc).
			WithGuardianGate(server.RequireRole(domain.RoleOwner, domain.RoleAdmin))
	}
	transferHandler := transfer.NewHandler(transferSvc).WithIdempotency(idemMW)
	fxHandler := fx.NewHandler(fxSvc).WithIdempotency(idemMW)
	routingHandler := routing.NewHandler()
	// Register payment routes for routing engine
	routingHandler.RegisterRoute(routing.NewStripeBankRoute(cfg.StripeSecretKey))
	// Create XDC client for blockchain route
	var xdcRouteClient *xdc.Client
	xdcRouteClient, err = xdc.New(context.Background(), cfg.XDCRPCURL, cfg.XDCChainID)
	if err != nil {
		log.Warn().Err(err).Msg("xdc bridge route: failed to init chain client")
	}
	// Treasury wallet sends TXDC; second wallet is demo recipient
	recipientAddr := "0x7c42b69b8668504cbbcd3dc9506f0493f4f75fed"
	routingHandler.RegisterRoute(routing.NewXDCBridgeRoute(xdcRouteClient, cfg.XDCTreasurySecretKey, recipientAddr))
	routingHandler.RegisterRoute(fiat.NewPaymentNetworkRoute("INR-EUR"))
	routingHandler.RegisterRoute(fiat.NewPaymentNetworkRoute("EUR-INR"))
	fiatHandler := fiat.NewHandler(fiatSvc)
	anchorFiatHandler := fiat.NewAnchorHandler(anchorFiatSvc)
	anchorHandler := anchor.NewHandler(anchorRegistry)
	feeHandler := fees.NewHandler(feeSvc)
	apikeyHandler := apikey.NewHandler(apiKeyRepo)
	webhookHandler := webhook.NewHandler(webhookSvc)
	batchHandler := batch.NewHandler(batchSvc).WithIdempotency(idemMW)
	scheduleHandler := schedule.NewHandler(scheduleSvc)
	treasuryHandler := treasury.NewHandler(treasurySvc).WithMutationGate(server.RequireRole(domain.RoleOwner, domain.RoleAdmin))

	srv := server.New(
		authHandler, orgHandler, walletHandler, transferHandler, fxHandler, fiatHandler,
		anchorFiatHandler, anchorHandler,
		feeHandler, reconcileHandler, apikeyHandler, apiKeyRepo,
		webhookHandler, batchHandler, scheduleHandler, treasuryHandler, complianceHandler, routingHandler, jwtSecretBytes, cfg.Port,
		map[string]server.DependencyCheck{
			"postgres": db.Ping,
			"replica":  func(ctx context.Context) error { return repoDB.ReplicaAvailable(ctx) },
			"redis":    func(ctx context.Context) error { return redisClient.Ping(ctx).Err() },

			"horizon": server.HorizonDependencyCheck(cfg.StellarHorizonURL),
			"worker": func(ctx context.Context) error {
				if _, err := redisClient.Get(ctx, "fluxa:worker:heartbeat").Result(); err != nil {
					return err
				}
				return nil
			},
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

	cancel() // stop the indexer's live payment stream

	asynqSrv.Shutdown()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}

	log.Info().Msg("goodbye")
}
