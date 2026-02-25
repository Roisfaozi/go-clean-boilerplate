package config

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role"
	roleRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user"
	userUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/router"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/handlers"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/circuitbreaker"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/storage"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tus"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	ws2 "github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Application holds all major application components.
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

// NewApplication initializes and wires up all application components.
func NewApplication(cfg *AppConfig) (*Application, error) {
	logger := NewLogrus(cfg)

	// Configure Circuit Breaker
	circuitbreaker.Configure(
		cfg.CircuitBreaker.Enabled,
		cfg.CircuitBreaker.MaxRequests,
		cfg.CircuitBreaker.Interval,
		cfg.CircuitBreaker.Timeout,
	)

	// Initialize OpenTelemetry
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

	// Redis Option for Asynq
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

	// Online Presence Manager
	presenceManager := ws2.NewPresenceManager(redisClient, logger, 5*time.Minute)

	// Ticket Manager
	ticketManager := ws2.NewRedisTicketManager(redisClient, 30*time.Second)

	wsConfig := NewDefaultWebSocketConfig()
	wsManager := ws2.NewWebSocketManager(wsConfig.ToPkgConfig(), logger, redisClient, presenceManager)
	go wsManager.Run()

	// Pruning Loop for Presence
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			removed, err := presenceManager.PruneStaleUsers(context.Background(), 1*time.Minute)
			if err != nil {
				logger.WithError(err).Error("Failed to prune stale users")
				continue
			}
			// Broadcast leave event for each pruned user
			for orgID, userIDs := range removed {
				for _, uid := range userIDs {
					wsManager.PresenceUpdate(orgID, "leave", &ws2.PresenceUser{UserID: uid})
				}
			}
		}
	}()

	logger.Info("Shared dependencies initialized.")

	sseManager := sse.NewManager()
	logger.Info("SSE Manager initialized.")

	globalEnforcer, err := NewCasbinEnforcer(cfg, dbConnection, logger)
	if err != nil {
		logger.Errorf("Error initializing casbin enforcer: %v", err)
		return nil, err
	}

	// Use TransactionalEnforcer to support SQL transaction integrity for policy changes
	enforcer := usecase.NewTransactionalEnforcer(globalEnforcer, cfg.Casbin.Model)

	storageProvider, err := NewStorageProvider(cfg)
	if err != nil {
		logger.Fatalf("Failed to initialize storage provider: %v", err)
	}
	logger.Infof("Storage provider initialized: %s", cfg.Storage.Driver)

	roleRepo := roleRepository.NewRoleRepository(dbConnection, logger)
	organizationRepository := orgRepo.NewOrganizationRepository(dbConnection)

	// Audit Module (Initialize early to inject into others)
	auditModule := audit.NewAuditModule(dbConnection, logger, validate, wsManager)

	// Inject TaskDistributor to AuthModule
	authModule := auth.NewAuthModule(
		cfg.Security.MaxLoginAttempts,
		cfg.Security.LockoutDuration,
		jwtManager,
		dbConnection,
		redisClient,
		logger,
		validate,
		tm,
		wsManager,
		sseManager,
		enforcer,
		auditModule,
		taskDistributor,
		organizationRepository,
		ticketManager,
		cfg.Casbin.DefaultRole,
		cfg.Casbin.DefaultDomain,
	)

	userModule := user.NewUserModule(dbConnection, logger, validate, tm, enforcer, auditModule, authModule, storageProvider)

	permissionModule := permission.NewPermissionModule(enforcer, validate, logger, roleRepo, userModule.UserRepo)

	roleModule := role.NewRoleModule(dbConnection, logger, validate, tm)

	accessModule := access.NewAccessModule(dbConnection, logger, validate)

	statsModule := stats.NewStatsModule(dbConnection, logger)

	projectModule := project.NewProjectModule(dbConnection, validate)

	organizationModule := organization.NewOrganizationModule(dbConnection, redisClient, taskDistributor, userModule.UserRepo, logger, validate, tm, enforcer, presenceManager, cfg.Server.FrontendBaseURL)

	logger.Info("Application modules initialized.")

	// Worker Handlers
	cleanupHandler := handlers.NewCleanupTaskHandler(
		authModule.TokenRepo,
		userModule.UserRepo,
		auditModule.AuditRepo,
		logger,
	)

	// Map AppConfig to WorkerConfig (Manual mapping to avoid cycle)
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

	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, logger, cleanupHandler, auditModule.AuditController.UseCase, auditModule.AuditRepo, workerCfg)
	scheduler := worker.NewScheduler(redisOpt, logger)
	scheduler.RegisterScheduledTasks()

	// Access AuthUseCase via AuthController
	authUseCase := authModule.AuthController.AuthUseCase
	authMiddleware := middleware.NewAuthMiddleware(authUseCase, logger, ticketManager)
	casbinMiddleware := middleware.CasbinMiddleware(enforcer, logger)
	tenantMiddleware := middleware.NewTenantMiddleware(
		organizationModule.OrgRepo,
		organizationModule.Reader(),
		logger,
	)
	wsController := ws2.NewWebSocketController(logger, wsManager, cfg.CORS.AllowedOrigins, userModule.UserRepo, enforcer)
	logger.Info("Middleware initialized.")

	// ---------------------------------------------------------
	// TUS Initialization
	// ---------------------------------------------------------
	tusRegistry := tus.NewRegistry()

	// Register Avatar Hook
	tusRegistry.Register("avatar", &userUseCase.AvatarHook{UserUseCase: userModule.UserUseCase})

	// Initialize AWS S3 Client for TUS
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.Storage.S3.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.Storage.S3.AccessKey, cfg.Storage.S3.SecretKey, "")),
	)
	if err != nil {
		logger.Errorf("Failed to load AWS config for TUS: %v", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.Storage.S3.ForcePathStyle
		if cfg.Storage.S3.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Storage.S3.Endpoint)
		}
	})

	tusHandler, err := tus.NewHandler(tus.Config{
		StorageDriver: cfg.Storage.Driver,
		LocalRootPath: cfg.Storage.Local.RootPath,
		S3Bucket:      cfg.Storage.S3.Bucket,
		S3Endpoint:    cfg.Storage.S3.Endpoint,
		BasePath:      cfg.Tus.BasePath,
	}, tusRegistry, s3Client, logger)
	if err != nil {
		logger.Errorf("Failed to init TUS handler: %v", err)
		// Don't kill app, just log? Or kill.
		// If TUS is critical, kill. But better to log for now.
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
		authModule,
		userModule,
		permissionModule,
		accessModule,
		roleModule,
		organizationModule,
		auditModule,
		statsModule,
		projectModule,
		authMiddleware,
		casbinMiddleware,
		tenantMiddleware,
		wsController,
		sseManager,
		dbConnection,
		redisClient,
		tusHandler, // Pass TUS Handler
		logger,
	)
	logger.Info("Router setup complete.")

	serverPort := fmt.Sprintf(":%d", cfg.Server.Port)
	httpServer := &http.Server{
		Addr:    serverPort,
		Handler: ginRouter,
	}
	logger.Infof("Server configured to run on port %s", serverPort)

	// Start Worker Processor in Goroutine
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

// Shutdown gracefully shuts down all application components.
func (app *Application) Shutdown(ctx context.Context) error {
	app.Log.Info("Shutting down HTTP server...")
	if err := app.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	// Shutdown Tracer
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
