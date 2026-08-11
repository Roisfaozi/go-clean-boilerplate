package config

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/router"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/handlers"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/circuitbreaker"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/storage"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	ws2 "github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	defaultPresencePruneInterval = 5 * time.Minute
	defaultTicketTTL             = 30 * time.Second
)

type Application struct {
	Server          *http.Server
	DB              *gorm.DB
	Redis           *redis.Client
	Log             *logrus.Logger
	Enforcer        permission.IEnforcer
	TaskDistributor worker.TaskDistributor
	TaskProcessor   worker.TaskProcessor
	Scheduler       *worker.Scheduler
	TracerShutdown  func(context.Context) error
	StorageProvider storage.Provider
}

func NewApplication(cfg *AppConfig) (*Application, error) {
	logger := NewLogrus(cfg)

	circuitbreaker.Configure(
		cfg.CircuitBreaker.Enabled,
		cfg.CircuitBreaker.MaxRequests,
		cfg.CircuitBreaker.Interval,
		cfg.CircuitBreaker.Timeout,
	)

	var tracerShutdown func(context.Context) error
	if cfg.Telemetry.Enabled {
		var err error
		tracerShutdown, err = telemetry.InitTracer(cfg.Telemetry.ServiceName, cfg.Telemetry.CollectorURL)
		if err != nil {
			logger.Errorf("Failed to initialize OTEL: %v", err)
		} else {
			logger.Infof("OTEL initialized for service: %s", cfg.Telemetry.ServiceName)
		}
	}

	validate := NewValidator()
	dbConnection := NewDatabase(cfg, logger)

	redisClient := NewRedisConfig(cfg, logger)

	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	tm := tx.NewTransactionManager(dbConnection, logger)

	jwtManager := jwt.NewJWTManager(
		cfg.JWT.AccessTokenSecret,
		cfg.JWT.RefreshTokenSecret,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
	)

	presenceManager := ws2.NewPresenceManager(redisClient, logger, defaultPresencePruneInterval)

	ticketManager := ws2.NewRedisTicketManager(redisClient, defaultTicketTTL)

	wsConfig := NewDefaultWebSocketConfig()
	wsManager := ws2.NewWebSocketManager(wsConfig.ToPkgConfig(), logger, redisClient, presenceManager)
	go wsManager.Run()

	logger.Info("Shared dependencies initialized.")

	sseManager := sse.NewManager()
	logger.Info("SSE Manager initialized.")

	globalEnforcer, err := NewCasbinEnforcer(cfg, dbConnection, logger)
	if err != nil {
		logger.Errorf("Error initializing casbin enforcer: %v", err)
		return nil, err
	}

	// ── Runtime Safety Guard ──
	if isStrictCasbinEnv(cfg.Server.AppEnv) {
		if globalEnforcer == nil {
			logger.Fatal("CRITICAL: Casbin is DISABLED outside local/test/dev. Set CASBIN_ENABLED=true. Aborting startup.")
		}
		policies, _ := globalEnforcer.GetPolicy()
		if len(policies) == 0 {
			logger.Fatal("CRITICAL: Casbin enforcer loaded with ZERO policies outside local/test/dev. Seed policies before deploying. Aborting startup.")
		}
		logger.Infof("Casbin strict environment guard passed: %d policies loaded.", len(policies))
	} else if globalEnforcer == nil {
		logger.Warn("Casbin is disabled. Authorization checks will be skipped. Only use this in local/test/dev.")
	}

	var enforcer usecase.IEnforcer
	if globalEnforcer != nil {
		enforcer = usecase.NewTransactionalEnforcer(globalEnforcer, cfg.Casbin.Model)
	}

	storageProvider, err := NewStorageProvider(cfg)
	if err != nil {
		logger.Fatalf("Failed to initialize storage provider: %v", err)
	}
	logger.Infof("Storage provider initialized: %s", cfg.Storage.Driver)

	ssoProviders := initSSOProviders(cfg)

	modules := initModules(
		cfg,
		dbConnection,
		redisClient,
		logger,
		validate,
		wsManager,
		presenceManager,
		taskDistributor,
		tm,
		jwtManager,
		ticketManager,
		enforcer,
		sseManager,
		storageProvider,
		ssoProviders,
	)
	logger.Info("Application modules initialized.")

	// Real-time Metrics Broadcaster & User Presence Pruner
	startMetricsBroadcaster(wsManager, presenceManager, modules.stats)

	cleanupHandler := handlers.NewCleanupTaskHandler(
		modules.auth.TokenRepo,
		modules.user.UserRepo,
		modules.audit.AuditRepo,
		logger,
	)

	webhookHandler := handlers.NewWebhookHandler(modules.webhook.Repo, logger)

	workerCfg := worker.WorkerConfig{
		SMTP: worker.SMTPConfig{
			Host:       cfg.SMTP.Host,
			Port:       cfg.SMTP.Port,
			Username:   cfg.SMTP.Username,
			Password:   cfg.SMTP.Password,
			FromSender: cfg.SMTP.FromSender,
			FromEmail:  cfg.SMTP.FromEmail,
		},
	}

	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, logger, cleanupHandler, webhookHandler, modules.audit.AuditController.UseCase, modules.audit.AuditRepo, workerCfg)
	scheduler := worker.NewScheduler(redisOpt, logger)
	scheduler.RegisterScheduledTasks()

	authUseCase := modules.auth.AuthController.AuthUseCase
	authMiddleware := middleware.NewAuthMiddleware(authUseCase, logger, ticketManager)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(modules.apiKey.UseCase, modules.user.UserRepo, logger)
	casbinMiddleware := middleware.CasbinMiddleware(enforcer, logger)
	tenantMiddleware := middleware.NewTenantMiddleware(
		modules.organization.OrgRepo,
		modules.organization.Reader(),
		logger,
	)
	wsController := ws2.NewWebSocketController(logger, wsManager, cfg.CORS.AllowedOrigins, modules.user.UserRepo, enforcer)
	logger.Info("Middleware initialized.")

	// TUS Initialization
	tusHandler, err := initTusHandler(cfg, logger, modules.user)
	if err != nil {
		logger.Errorf("Failed to init TUS handler: %v", err)
	} else {
		logger.Info("TUS Handler initialized.")
	}

	configRouter := router.RouterConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		TrustedProxies:   cfg.Server.TrustedProxies,
		RateLimitEnabled: cfg.RateLimit.Enabled,
		RateLimitRPS:     cfg.RateLimit.RPS,
		RateLimitBurst:   cfg.RateLimit.Burst,
		RateLimitStore:   cfg.RateLimit.Store,
		MetricsEnabled:   cfg.Metrics.Enabled,
		MetricsAuth:      cfg.Metrics.AuthEnabled,
		MetricsUser:      cfg.Metrics.Username,
		MetricsPass:      cfg.Metrics.Password,
		OTEL: struct {
			Enabled     bool
			ServiceName string
		}{
			Enabled:     cfg.Telemetry.Enabled,
			ServiceName: cfg.Telemetry.ServiceName,
		},
	}

	ginRouter := router.SetupRouter(
		configRouter,
		modules.auth,
		modules.user,
		modules.permission,
		modules.access,
		modules.role,
		modules.organization,
		modules.audit,
		modules.stats,
		modules.project,
		modules.apiKey,
		modules.webhook,
		authMiddleware,
		apiKeyMiddleware,
		casbinMiddleware,
		tenantMiddleware,
		wsController,
		sseManager,
		dbConnection,
		redisClient,
		tusHandler,
		logger,
	)
	logger.Info("Router setup complete.")

	serverPort := fmt.Sprintf(":%d", cfg.Server.Port)
	httpServer := &http.Server{
		Addr:    serverPort,
		Handler: ginRouter,
	}
	logger.Infof("Server configured to run on port %s", serverPort)

	go func() {
		logger.Info("Starting Background Worker Processor...")
		if err := taskProcessor.Start(); err != nil {
			logger.Fatalf("Failed to start worker processor: %v", err)
		}
	}()

	app := &Application{
		Server:          httpServer,
		DB:              dbConnection,
		Redis:           redisClient,
		Log:             logger,
		Enforcer:        enforcer,
		TaskDistributor: taskDistributor,
		TaskProcessor:   taskProcessor,
		Scheduler:       scheduler,
		TracerShutdown:  tracerShutdown,
		StorageProvider: storageProvider,
	}

	return app, nil
}

func isStrictCasbinEnv(appEnv string) bool {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "", defaultAppEnvLocal, defaultAppEnvDev, defaultAppEnvDevelopment, defaultAppEnvTest, defaultAppEnvTesting:
		return false
	default:
		return true
	}
}

func (app *Application) Shutdown(ctx context.Context) error {
	app.Log.Info("Shutting down HTTP server...")
	if err := app.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	if app.TracerShutdown != nil {
		app.Log.Info("Shutting down Tracer Provider...")
		if err := app.TracerShutdown(ctx); err != nil {
			app.Log.Errorf("Failed to shutdown Tracer: %v", err)
		}
	}

	app.Log.Info("Shutting down Worker Processor...")
	app.TaskProcessor.Shutdown()
	app.Scheduler.Shutdown()

	if app.Redis != nil {
		app.Log.Info("Closing Redis connection...")
		if err := app.Redis.Close(); err != nil {
			app.Log.Errorf("Failed to close Redis client: %v", err)
		}
	}

	if app.DB != nil {
		app.Log.Info("Closing database connection...")
		sqlDB, err := app.DB.DB()
		if err != nil {
			app.Log.Errorf("Failed to get DB instance for closing: %v", err)
		} else if err := sqlDB.Close(); err != nil {
			app.Log.Errorf("Failed to close database connection: %v", err)
		}
	}

	return nil
}
