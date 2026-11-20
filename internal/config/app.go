package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook"
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
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/tus/tusd/v2/pkg/handler"
	"go.uber.org/fx"
)

const (
	defaultPresencePruneInterval = 5 * time.Minute
	defaultTicketTTL             = 30 * time.Second
)

// TracerShutdownFunc flushes and shuts down the configured OpenTelemetry tracer provider.
type TracerShutdownFunc func(context.Context) error

type Application struct {
	Server            *http.Server
	DB                *sqlx.DB
	Redis             *redis.Client
	Log               *logrus.Logger
	Enforcer          permission.IEnforcer
	TaskDistributor   worker.TaskDistributor
	TaskProcessor     worker.TaskProcessor
	Scheduler         *worker.Scheduler
	TracerShutdown    func(context.Context) error
	StorageProvider   storage.Provider
	BroadcasterCancel context.CancelFunc

	wsManager   *ws2.WebSocketManager
	sseManager  *sse.Manager
	broadcaster *MetricsBroadcaster
	sseStopOnce sync.Once
}

func configureCircuitBreaker(cfg *AppConfig) {
	circuitbreaker.Configure(
		cfg.CircuitBreaker.Enabled,
		cfg.CircuitBreaker.MaxRequests,
		cfg.CircuitBreaker.Interval,
		cfg.CircuitBreaker.Timeout,
	)
}

func provideTracerShutdown(cfg *AppConfig, logger *logrus.Logger) TracerShutdownFunc {
	if !cfg.Telemetry.Enabled {
		return nil
	}

	tracerShutdown, err := telemetry.InitTracer(telemetry.Config{
		ServiceName:  cfg.Telemetry.ServiceName,
		CollectorURL: cfg.Telemetry.CollectorURL,
		Insecure:     cfg.Telemetry.Insecure,
		SampleRatio:  cfg.Telemetry.SampleRatio,
	})
	if err != nil {
		logger.Errorf("Failed to initialize OTEL: %v", err)
		return nil
	}

	logger.Infof("OTEL initialized for service: %s (insecure: %v, sample_ratio: %.2f)",
		cfg.Telemetry.ServiceName, cfg.Telemetry.Insecure, cfg.Telemetry.SampleRatio)
	return tracerShutdown
}

func provideAsynqRedisOpt(cfg *AppConfig) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
}

func provideTaskDistributor(redisOpt asynq.RedisClientOpt) worker.TaskDistributor {
	return worker.NewRedisTaskDistributor(redisOpt)
}

func provideTransactionManager(db *sqlx.DB, log *logrus.Logger) tx.WithTransactionManager {
	return tx.NewSQLXTransactionManager(db, log)
}

func provideJWTManager(cfg *AppConfig) *jwt.JWTManager {
	return jwt.NewJWTManager(
		cfg.JWT.AccessTokenSecret,
		cfg.JWT.RefreshTokenSecret,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
	)
}

func providePresenceManager(redisClient *redis.Client, logger *logrus.Logger) ws2.PresenceManager {
	return ws2.NewPresenceManager(redisClient, logger, defaultPresencePruneInterval)
}

func provideTicketManager(redisClient *redis.Client) ws2.TicketManager {
	return ws2.NewRedisTicketManager(redisClient, defaultTicketTTL)
}

func provideWebSocketManager(
	logger *logrus.Logger,
	redisClient *redis.Client,
	presenceManager ws2.PresenceManager,
) *ws2.WebSocketManager {
	wsConfig := NewDefaultWebSocketConfig()
	wsManager := ws2.NewWebSocketManager(wsConfig.ToPkgConfig(), logger, redisClient, presenceManager)
	logger.Info("Shared dependencies initialized.")
	return wsManager
}

func provideSSEManager(logger *logrus.Logger) *sse.Manager {
	sseManager := sse.NewManager()
	logger.Info("SSE Manager initialized.")
	return sseManager
}

func provideTransactionalEnforcer(
	cfg *AppConfig,
	sqlxDB *sqlx.DB,
	logger *logrus.Logger,
) (usecase.IEnforcer, error) {
	globalEnforcer, err := NewCasbinEnforcer(cfg, sqlxDB, logger)
	if err != nil {
		logger.Errorf("Error initializing casbin enforcer: %v", err)
		return nil, err
	}

	// ── Runtime Safety Guard ──
	if isStrictCasbinEnv(cfg.Server.AppEnv) {
		if globalEnforcer == nil {
			err := errors.New("CRITICAL: Casbin is DISABLED outside local/test/dev. Set CASBIN_ENABLED=true. Aborting startup")
			logger.Error(err)
			return nil, err
		}
		policies, err := globalEnforcer.GetPolicy()
		if err != nil {
			return nil, fmt.Errorf("failed to inspect casbin policies: %w", err)
		}
		if len(policies) == 0 {
			err := errors.New("CRITICAL: Casbin enforcer loaded with ZERO policies outside local/test/dev. Seed policies before deploying. Aborting startup")
			logger.Error(err)
			return nil, err
		}
		logger.Infof("Casbin strict environment guard passed: %d policies loaded.", len(policies))
	} else if globalEnforcer == nil {
		logger.Warn("Casbin is disabled. Authorization checks will be skipped. Only use this in local/test/dev.")
	}

	if globalEnforcer == nil {
		return nil, nil
	}

	return usecase.NewTransactionalEnforcer(globalEnforcer, cfg.Casbin.Model), nil
}

func provideStorageProvider(cfg *AppConfig, logger *logrus.Logger) (storage.Provider, error) {
	storageProvider, err := NewStorageProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage provider: %w", err)
	}
	logger.Infof("Storage provider initialized: %s", cfg.Storage.Driver)
	return storageProvider, nil
}

func registerPrometheusCollectors(
	cfg *AppConfig,
	sqlxDB *sqlx.DB,
	sseManager *sse.Manager,
	wsManager *ws2.WebSocketManager,
	logger *logrus.Logger,
) {
	registerPrometheusCollectorsWith(prometheus.DefaultRegisterer, cfg, sqlxDB, sseManager, wsManager, logger)
}

func registerPrometheusCollectorsWith(
	registerer prometheus.Registerer,
	cfg *AppConfig,
	sqlxDB *sqlx.DB,
	sseManager *sse.Manager,
	wsManager *ws2.WebSocketManager,
	logger *logrus.Logger,
) {
	if registerer == nil {
		return
	}

	var sqlDB *sql.DB
	if sqlxDB != nil {
		sqlDB = sqlxDB.DB
	}
	dbName := ""
	if cfg != nil {
		dbName = cfg.Mysql.DBName
	}

	var sseCountFn func() int
	if sseManager != nil {
		sseCountFn = sseManager.ClientCount
	}
	var wsCountFn func() int
	if wsManager != nil {
		wsCountFn = wsManager.ClientCount
	}

	registerCollectorWith(registerer, telemetry.NewDBStatsCollector(sqlDB, dbName), "db_stats", logger)
	registerCollectorWith(registerer, telemetry.NewRealtimeStatsCollector(sseCountFn, wsCountFn), "realtime_stats", logger)
}

func registerCollectorWith(
	registerer prometheus.Registerer,
	collector prometheus.Collector,
	name string,
	logger *logrus.Logger,
) {
	if err := registerer.Register(collector); err != nil {
		var alreadyRegistered prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyRegistered) {
			return
		}
		if logger != nil {
			logger.WithError(err).Warnf("Failed to register %s Prometheus collector", name)
		}
	}
}

func provideCleanupTaskHandler(
	authMod *auth.AuthModule,
	userMod *user.UserModule,
	auditMod *audit.AuditModule,
	logger *logrus.Logger,
) *handlers.CleanupTaskHandler {
	return handlers.NewCleanupTaskHandler(
		authMod.TokenRepo,
		userMod.UserRepo,
		auditMod.AuditRepo,
		logger,
	)
}

func provideWebhookHandler(
	webhookMod *webhook.WebhookModule,
	logger *logrus.Logger,
) *handlers.WebhookHandler {
	return handlers.NewWebhookHandler(webhookMod.Repo, logger)
}

func provideTaskProcessor(
	cfg *AppConfig,
	redisOpt asynq.RedisClientOpt,
	logger *logrus.Logger,
	cleanupHandler *handlers.CleanupTaskHandler,
	webhookHandler *handlers.WebhookHandler,
	auditMod *audit.AuditModule,
) worker.TaskProcessor {
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

	return worker.NewRedisTaskProcessor(
		redisOpt,
		logger,
		cleanupHandler,
		webhookHandler,
		auditMod.AuditController.UseCase,
		auditMod.AuditRepo,
		workerCfg,
	)
}

func provideScheduler(redisOpt asynq.RedisClientOpt, logger *logrus.Logger) *worker.Scheduler {
	scheduler := worker.NewScheduler(redisOpt, logger)
	scheduler.RegisterScheduledTasks()
	return scheduler
}

func provideAuthMiddleware(
	authMod *auth.AuthModule,
	logger *logrus.Logger,
	ticketManager ws2.TicketManager,
) *middleware.AuthMiddleware {
	return middleware.NewAuthMiddleware(authMod.AuthController.AuthUseCase, logger, ticketManager)
}

func provideAPIKeyMiddleware(
	apiKeyMod *api_key.ApiKeyModule,
	userMod *user.UserModule,
	logger *logrus.Logger,
) *middleware.APIKeyMiddleware {
	return middleware.NewAPIKeyMiddleware(apiKeyMod.UseCase, userMod.UserRepo, logger)
}

func provideTenantMiddleware(
	orgMod *organization.OrganizationModule,
	logger *logrus.Logger,
) *middleware.TenantMiddleware {
	return middleware.NewTenantMiddleware(
		orgMod.OrgRepo,
		orgMod.Reader(),
		logger,
	)
}

func provideWebSocketController(
	cfg *AppConfig,
	logger *logrus.Logger,
	wsManager *ws2.WebSocketManager,
	userMod *user.UserModule,
	enforcer usecase.IEnforcer,
) *ws2.WebSocketController {
	return ws2.NewWebSocketController(logger, wsManager, cfg.CORS.AllowedOrigins, userMod.UserRepo, enforcer)
}

type routerParams struct {
	fx.In

	Cfg              *AppConfig
	Auth             *auth.AuthModule
	User             *user.UserModule
	Permission       *permission.PermissionModule
	Access           *access.AccessModule
	Role             *role.RoleModule
	Organization     *organization.OrganizationModule
	Audit            *audit.AuditModule
	Stats            *stats.StatsModule
	Project          *project.ProjectModule
	APIKey           *api_key.ApiKeyModule
	Webhook          *webhook.WebhookModule
	AuthMiddleware   *middleware.AuthMiddleware
	APIKeyMiddleware *middleware.APIKeyMiddleware
	Enforcer         usecase.IEnforcer
	TenantMiddleware *middleware.TenantMiddleware
	WSController     *ws2.WebSocketController
	SSEManager       *sse.Manager
	DB               *sqlx.DB
	Redis            *redis.Client
	TusHandler       *handler.Handler
	Logger           *logrus.Logger
}

func provideStdRouter(p routerParams) http.Handler {
	stdRouterCfg := router.StdRouterConfig{
		AllowedOrigins:   p.Cfg.CORS.AllowedOrigins,
		TrustedProxies:   p.Cfg.Server.TrustedProxies,
		RateLimitEnabled: p.Cfg.RateLimit.Enabled,
		RateLimitRPS:     p.Cfg.RateLimit.RPS,
		RateLimitBurst:   p.Cfg.RateLimit.Burst,
		RateLimitStore:   p.Cfg.RateLimit.Store,
		MetricsEnabled:   p.Cfg.Metrics.Enabled,
		MetricsAuth:      p.Cfg.Metrics.AuthEnabled,
		MetricsUser:      p.Cfg.Metrics.Username,
		MetricsPass:      p.Cfg.Metrics.Password,
		OTEL: struct {
			Enabled     bool
			ServiceName string
		}{
			Enabled:     p.Cfg.Telemetry.Enabled,
			ServiceName: p.Cfg.Telemetry.ServiceName,
		},
	}

	stdRouter := router.SetupStdRouter(
		stdRouterCfg,
		p.Auth,
		p.User,
		p.Permission,
		p.Access,
		p.Role,
		p.Organization,
		p.Audit,
		p.Stats,
		p.Project,
		p.APIKey,
		p.Webhook,
		p.AuthMiddleware,
		p.APIKeyMiddleware,
		p.Enforcer,
		p.TenantMiddleware,
		p.WSController,
		p.SSEManager,
		p.DB,
		p.Redis,
		p.TusHandler,
		p.Logger,
	)
	p.Logger.Info("Stdlib router setup complete.")
	return stdRouter
}

func provideHTTPServer(cfg *AppConfig, stdRouter http.Handler, logger *logrus.Logger) *http.Server {
	serverPort := fmt.Sprintf(":%d", cfg.Server.Port)
	httpServer := &http.Server{
		Addr:              serverPort,
		Handler:           stdRouter,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       120 * time.Second,
	}
	logger.Infof("Server configured to run on port %s", serverPort)
	return httpServer
}

type applicationParams struct {
	fx.In

	Server          *http.Server
	DB              *sqlx.DB
	Redis           *redis.Client
	Log             *logrus.Logger
	Enforcer        usecase.IEnforcer
	TaskDistributor worker.TaskDistributor
	TaskProcessor   worker.TaskProcessor
	Scheduler       *worker.Scheduler
	TracerShutdown  TracerShutdownFunc
	StorageProvider storage.Provider
	WSManager       *ws2.WebSocketManager
	SSEManager      *sse.Manager
	Broadcaster     *MetricsBroadcaster
}

func provideApplication(p applicationParams) *Application {
	return &Application{
		Server:          p.Server,
		DB:              p.DB,
		Redis:           p.Redis,
		Log:             p.Log,
		Enforcer:        p.Enforcer,
		TaskDistributor: p.TaskDistributor,
		TaskProcessor:   p.TaskProcessor,
		Scheduler:       p.Scheduler,
		TracerShutdown:  p.TracerShutdown,
		StorageProvider: p.StorageProvider,
		wsManager:       p.WSManager,
		sseManager:      p.SSEManager,
		broadcaster:     p.Broadcaster,
	}
}

// NewApplication constructs an Application using the same shared constructors as CoreFxModule
// and starts background loops expected by direct callers (such as E2E tests).
func NewApplication(cfg *AppConfig) (*Application, error) {
	logger := NewLogrus(cfg)
	configureCircuitBreaker(cfg)
	tracerShutdown := provideTracerShutdown(cfg, logger)
	validate := NewValidator()
	sqlxDB := NewSQLXDatabase(cfg, logger)
	redisClient := NewRedisConfig(cfg, logger)
	redisOpt := provideAsynqRedisOpt(cfg)
	taskDistributor := provideTaskDistributor(redisOpt)
	tm := provideTransactionManager(sqlxDB, logger)
	jwtManager := provideJWTManager(cfg)
	presenceManager := providePresenceManager(redisClient, logger)
	ticketManager := provideTicketManager(redisClient)
	wsManager := provideWebSocketManager(logger, redisClient, presenceManager)
	sseManager := provideSSEManager(logger)

	registerPrometheusCollectors(cfg, sqlxDB, sseManager, wsManager, logger)

	enforcer, err := provideTransactionalEnforcer(cfg, sqlxDB, logger)
	if err != nil {
		return nil, err
	}

	storageProvider, err := provideStorageProvider(cfg, logger)
	if err != nil {
		return nil, err
	}

	ssoProviders := initSSOProviders(cfg)

	roleRepo := provideRoleRepository(sqlxDB, logger)
	organizationRepo := provideOrganizationRepository(sqlxDB, redisClient)
	auditMod := provideAuditModule(sqlxDB, logger, validate, wsManager, taskDistributor)
	authMod := provideAuthModule(authModuleParams{
		Cfg:             cfg,
		JWTManager:      jwtManager,
		DB:              sqlxDB,
		Redis:           redisClient,
		Logger:          logger,
		Validate:        validate,
		TM:              tm,
		WSManager:       wsManager,
		SSEManager:      sseManager,
		Enforcer:        enforcer,
		AuditMod:        auditMod,
		TaskDistributor: taskDistributor,
		OrgRepo:         organizationRepo,
		TicketManager:   ticketManager,
		SSOProviders:    ssoProviders,
	})
	webhookMod := provideWebhookModule(sqlxDB, logger, validate, taskDistributor)
	userMod := provideUserModule(userModuleParams{
		DB:              sqlxDB,
		Logger:          logger,
		Validate:        validate,
		TM:              tm,
		Enforcer:        enforcer,
		AuditMod:        auditMod,
		AuthMod:         authMod,
		WebhookMod:      webhookMod,
		StorageProvider: storageProvider,
	})
	apiKeyMod := provideAPIKeyModule(sqlxDB, userMod, redisClient, logger, validate)
	accessMod := provideAccessModule(sqlxDB, logger, validate)
	permissionMod := providePermissionModule(enforcer, validate, logger, roleRepo, userMod, accessMod, auditMod)
	roleMod := provideRoleModule(sqlxDB, logger, validate, tm, permissionMod)
	statsMod := provideStatsModule(sqlxDB, logger)
	projectMod := provideProjectModule(sqlxDB, validate)
	orgMod := provideOrganizationModule(organizationModuleParams{
		Cfg:             cfg,
		DB:              sqlxDB,
		Redis:           redisClient,
		TaskDistributor: taskDistributor,
		UserMod:         userMod,
		Logger:          logger,
		Validate:        validate,
		TM:              tm,
		Enforcer:        enforcer,
		PresenceManager: presenceManager,
		RoleRepo:        roleRepo,
	})
	logger.Info("Application modules initialized.")

	broadcaster := provideMetricsBroadcaster(wsManager, presenceManager, statsMod)
	cleanupHandler := provideCleanupTaskHandler(authMod, userMod, auditMod, logger)
	webhookHandler := provideWebhookHandler(webhookMod, logger)
	taskProcessor := provideTaskProcessor(cfg, redisOpt, logger, cleanupHandler, webhookHandler, auditMod)
	scheduler := provideScheduler(redisOpt, logger)

	authMiddleware := provideAuthMiddleware(authMod, logger, ticketManager)
	apiKeyMiddleware := provideAPIKeyMiddleware(apiKeyMod, userMod, logger)
	tenantMiddleware := provideTenantMiddleware(orgMod, logger)
	wsController := provideWebSocketController(cfg, logger, wsManager, userMod, enforcer)
	logger.Info("Middleware initialized.")

	tusHandler := provideTusHandler(cfg, logger, userMod)

	stdRouter := provideStdRouter(routerParams{
		Cfg:              cfg,
		Auth:             authMod,
		User:             userMod,
		Permission:       permissionMod,
		Access:           accessMod,
		Role:             roleMod,
		Organization:     orgMod,
		Audit:            auditMod,
		Stats:            statsMod,
		Project:          projectMod,
		APIKey:           apiKeyMod,
		Webhook:          webhookMod,
		AuthMiddleware:   authMiddleware,
		APIKeyMiddleware: apiKeyMiddleware,
		Enforcer:         enforcer,
		TenantMiddleware: tenantMiddleware,
		WSController:     wsController,
		SSEManager:       sseManager,
		DB:               sqlxDB,
		Redis:            redisClient,
		TusHandler:       tusHandler,
		Logger:           logger,
	})
	httpServer := provideHTTPServer(cfg, stdRouter, logger)

	app := provideApplication(applicationParams{
		Server:          httpServer,
		DB:              sqlxDB,
		Redis:           redisClient,
		Log:             logger,
		Enforcer:        enforcer,
		TaskDistributor: taskDistributor,
		TaskProcessor:   taskProcessor,
		Scheduler:       scheduler,
		TracerShutdown:  tracerShutdown,
		StorageProvider: storageProvider,
		WSManager:       wsManager,
		SSEManager:      sseManager,
		Broadcaster:     broadcaster,
	})

	go wsManager.Run()
	app.BroadcasterCancel = broadcaster.Start()
	go func() {
		logger.Info("Starting Background Worker Processor...")
		if err := taskProcessor.Start(); err != nil {
			logger.Errorf("Failed to start worker processor: %v", err)
		}
	}()

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

func (app *Application) logInfo(msg string) {
	if app.Log != nil {
		app.Log.Info(msg)
	}
}

func (app *Application) logErrorf(format string, args ...interface{}) {
	if app.Log != nil {
		app.Log.Errorf(format, args...)
	}
}

func (app *Application) stopBroadcaster() {
	if app.broadcaster != nil {
		app.logInfo("Stopping Real-time Metrics Broadcaster & Presence Pruner...")
		app.broadcaster.Stop()
		return
	}
	if app.BroadcasterCancel != nil {
		app.logInfo("Stopping Real-time Metrics Broadcaster & Presence Pruner...")
		app.BroadcasterCancel()
	}
}

func (app *Application) stopSSEManager() {
	if app.sseManager == nil {
		return
	}
	app.sseStopOnce.Do(func() {
		app.logInfo("Stopping SSE Manager...")
		app.sseManager.Stop()
	})
}

// Shutdown gracefully stops the application in strict shutdown order:
// HTTP server -> scheduler/worker intake -> broadcaster/pruner & realtime managers -> Redis/SQL -> tracer.
func (app *Application) Shutdown(ctx context.Context) error {
	var errs []error

	// 1. Stop accepting HTTP requests
	if app.Server != nil {
		app.logInfo("Shutting down HTTP server...")
		if err := app.Server.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("server forced to shutdown: %w", err))
		}
	}

	// 2. Stop scheduler and worker intake
	if app.Scheduler != nil {
		app.logInfo("Shutting down Scheduler...")
		app.Scheduler.Shutdown()
	}
	if app.TaskProcessor != nil {
		app.logInfo("Shutting down Worker Processor...")
		app.TaskProcessor.Shutdown()
	}

	// 3. Stop broadcaster/pruner & close realtime managers
	app.stopBroadcaster()
	if app.wsManager != nil {
		app.logInfo("Stopping WebSocket Manager...")
		app.wsManager.Stop()
	}
	app.stopSSEManager()

	// 4. Close Redis & SQL connections
	if app.Redis != nil {
		app.logInfo("Closing Redis connection...")
		if err := app.Redis.Close(); err != nil {
			app.logErrorf("Failed to close Redis client: %v", err)
			errs = append(errs, err)
		}
	}
	if app.DB != nil {
		app.logInfo("Closing database connection...")
		if err := app.DB.Close(); err != nil {
			app.logErrorf("Failed to close database connection: %v", err)
			errs = append(errs, err)
		}
	}

	// 5. Shutdown tracer provider
	if app.TracerShutdown != nil {
		app.logInfo("Shutting down Tracer Provider...")
		if err := app.TracerShutdown(ctx); err != nil {
			app.logErrorf("Failed to shutdown Tracer: %v", err)
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
