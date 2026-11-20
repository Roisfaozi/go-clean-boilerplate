package config

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

// InfraFxModule provides shared infrastructure singletons: logger, validator, tracer,
// SQL DB, Redis, transaction manager, JWT, storage, Casbin enforcer, realtime managers,
// SSO providers, and task distributor, plus circuit breaker and Prometheus collector registration.
var InfraFxModule = fx.Module("infra",
	fx.Provide(
		NewLogrus,
		NewValidator,
		provideTracerShutdown,
		NewSQLXDatabase,
		NewRedisConfig,
		provideAsynqRedisOpt,
		provideTaskDistributor,
		provideTransactionManager,
		provideJWTManager,
		providePresenceManager,
		provideTicketManager,
		provideWebSocketManager,
		provideSSEManager,
		provideTransactionalEnforcer,
		provideStorageProvider,
		initSSOProviders,
	),
	fx.Invoke(
		configureCircuitBreaker,
		registerPrometheusCollectors,
	),
)

// AppModulesFxModule provides shared cross-module repositories and all 11 domain modules.
var AppModulesFxModule = fx.Module("modules",
	fx.Provide(
		provideRoleRepository,
		provideOrganizationRepository,
		provideAuditModule,
		provideAuthModule,
		provideWebhookModule,
		provideUserModule,
		provideAPIKeyModule,
		provideAccessModule,
		providePermissionModule,
		provideRoleModule,
		provideStatsModule,
		provideProjectModule,
		provideOrganizationModule,
	),
)

// RuntimeFxModule provides background workers, scheduler, middlewares, TUS handler,
// HTTP router, *http.Server, and *Application, and registers lifecycle hooks.
var RuntimeFxModule = fx.Module("runtime",
	fx.Provide(
		provideMetricsBroadcaster,
		provideCleanupTaskHandler,
		provideWebhookHandler,
		provideTaskProcessor,
		provideScheduler,
		provideAuthMiddleware,
		provideAPIKeyMiddleware,
		provideTenantMiddleware,
		provideWebSocketController,
		provideTusHandler,
		provideStdRouter,
		provideHTTPServer,
		provideApplication,
	),
	fx.Invoke(RegisterAppLifecycle),
)

// CoreFxModule is the top-level Fx composition root for the API server.
var CoreFxModule = fx.Module("core",
	InfraFxModule,
	AppModulesFxModule,
	RuntimeFxModule,
)

type lifecycleHooks struct {
	StartHTTP        func()
	StopHTTP         func(context.Context) error
	StartScheduler   func()
	StopScheduler    func()
	StartProcessor   func()
	StopProcessor    func()
	StartBroadcaster func()
	StopBroadcaster  func()
	StartWS          func()
	StopWS           func()
	StopSSE          func()
	CloseRedis       func() error
	CloseDB          func() error
	ShutdownTracer   func(context.Context) error
}

// registerLifecycleHooks appends Fx hooks in startup order (FIFO) so that Fx
// executes OnStop hooks in strict reverse (LIFO) shutdown order:
// HTTP -> Scheduler -> Worker Processor -> Broadcaster/Pruner -> WS Manager -> SSE Manager -> Redis -> SQL DB -> Tracer.
func registerLifecycleHooks(lc fx.Lifecycle, hooks lifecycleHooks) {
	// 1. Tracer provider (stops LAST)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if hooks.ShutdownTracer != nil {
				return hooks.ShutdownTracer(ctx)
			}
			return nil
		},
	})

	// 2. Database & Redis connections (stops 4th: Redis then SQL DB)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if hooks.CloseDB != nil {
				return hooks.CloseDB()
			}
			return nil
		},
	})
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if hooks.CloseRedis != nil {
				return hooks.CloseRedis()
			}
			return nil
		},
	})

	// 3. Realtime managers & metrics broadcaster/pruner (stops 3rd: broadcaster -> WS -> SSE)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if hooks.StopSSE != nil {
				hooks.StopSSE()
			}
			return nil
		},
	})
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if hooks.StartWS != nil {
				hooks.StartWS()
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if hooks.StopWS != nil {
				hooks.StopWS()
			}
			return nil
		},
	})
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if hooks.StartBroadcaster != nil {
				hooks.StartBroadcaster()
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if hooks.StopBroadcaster != nil {
				hooks.StopBroadcaster()
			}
			return nil
		},
	})

	// 4. Worker processor & scheduler (stops 2nd: scheduler -> worker processor)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if hooks.StartProcessor != nil {
				hooks.StartProcessor()
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if hooks.StopProcessor != nil {
				hooks.StopProcessor()
			}
			return nil
		},
	})
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if hooks.StartScheduler != nil {
				hooks.StartScheduler()
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if hooks.StopScheduler != nil {
				hooks.StopScheduler()
			}
			return nil
		},
	})

	// 5. HTTP server (starts LAST, stops FIRST)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if hooks.StartHTTP != nil {
				hooks.StartHTTP()
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if hooks.StopHTTP != nil {
				return hooks.StopHTTP(ctx)
			}
			return nil
		},
	})
}

// RegisterAppLifecycle registers ordered Fx lifecycle hooks for graceful startup and shutdown.
func RegisterAppLifecycle(lc fx.Lifecycle, app *Application, log *logrus.Logger) {
	if log == nil && app != nil {
		log = app.Log
	}

	logInfo := func(msg string) {
		if log != nil {
			log.Info(msg)
		}
	}
	logInfof := func(format string, args ...interface{}) {
		if log != nil {
			log.Infof(format, args...)
		}
	}
	logErrorf := func(format string, args ...interface{}) {
		if log != nil {
			log.Errorf(format, args...)
		}
	}

	registerLifecycleHooks(lc, lifecycleHooks{
		StartHTTP: func() {
			if app == nil || app.Server == nil {
				return
			}
			go func() {
				logInfof("Server starting on %s", app.Server.Addr)
				if err := app.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logErrorf("Server error: %v", err)
				}
			}()
		},
		StopHTTP: func(ctx context.Context) error {
			if app == nil || app.Server == nil {
				return nil
			}
			logInfo("Shutting down HTTP server...")
			if err := app.Server.Shutdown(ctx); err != nil {
				return fmt.Errorf("server forced to shutdown: %w", err)
			}
			return nil
		},
		StartScheduler: func() {
			if app == nil || app.Scheduler == nil {
				return
			}
			go func() {
				logInfo("Starting Scheduler...")
				if err := app.Scheduler.Start(); err != nil {
					logErrorf("Failed to start scheduler: %v", err)
				}
			}()
		},
		StopScheduler: func() {
			if app == nil || app.Scheduler == nil {
				return
			}
			logInfo("Shutting down Scheduler...")
			app.Scheduler.Shutdown()
		},
		StartProcessor: func() {
			if app == nil || app.TaskProcessor == nil {
				return
			}
			go func() {
				logInfo("Starting Background Worker Processor...")
				if err := app.TaskProcessor.Start(); err != nil {
					logErrorf("Failed to start worker processor: %v", err)
				}
			}()
		},
		StopProcessor: func() {
			if app == nil || app.TaskProcessor == nil {
				return
			}
			logInfo("Shutting down Worker Processor...")
			app.TaskProcessor.Shutdown()
		},
		StartBroadcaster: func() {
			if app == nil || app.broadcaster == nil {
				return
			}
			app.BroadcasterCancel = app.broadcaster.Start()
		},
		StopBroadcaster: func() {
			if app == nil {
				return
			}
			app.stopBroadcaster()
		},
		StartWS: func() {
			if app == nil || app.wsManager == nil {
				return
			}
			go app.wsManager.Run()
		},
		StopWS: func() {
			if app == nil || app.wsManager == nil {
				return
			}
			logInfo("Stopping WebSocket Manager...")
			app.wsManager.Stop()
		},
		StopSSE: func() {
			if app == nil {
				return
			}
			app.stopSSEManager()
		},
		CloseRedis: func() error {
			if app == nil || app.Redis == nil {
				return nil
			}
			logInfo("Closing Redis connection...")
			if err := app.Redis.Close(); err != nil {
				logErrorf("Failed to close Redis client: %v", err)
				return err
			}
			return nil
		},
		CloseDB: func() error {
			if app == nil || app.DB == nil {
				return nil
			}
			logInfo("Closing database connection...")
			if err := app.DB.Close(); err != nil {
				logErrorf("Failed to close database connection: %v", err)
				return err
			}
			return nil
		},
		ShutdownTracer: func(ctx context.Context) error {
			if app == nil || app.TracerShutdown == nil {
				return nil
			}
			logInfo("Shutting down Tracer Provider...")
			if err := app.TracerShutdown(ctx); err != nil {
				logErrorf("Failed to shutdown Tracer: %v", err)
				return err
			}
			return nil
		},
	})
}
