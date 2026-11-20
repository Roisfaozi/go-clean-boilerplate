package config

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth"
	authHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/delivery/http"
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
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sso"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/storage"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tus"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	ws2 "github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/tus/tusd/v2/pkg/handler"
	"go.uber.org/fx"
)

const (
	providerGoogle    = "google"
	providerMicrosoft = "microsoft"
	providerGitHub    = "github"
	metricsChannel    = "system:metrics"
	metricsEventType  = "metrics_update"
)

func initSSOProviders(cfg *AppConfig) map[string]sso.Provider {
	ssoProviders := make(map[string]sso.Provider)
	ssoProviders[providerGoogle] = sso.NewGoogleProvider(sso.ProviderConfig{
		ClientID:     cfg.SSO.Google.ClientID,
		ClientSecret: cfg.SSO.Google.ClientSecret,
		RedirectURL:  cfg.SSO.Google.RedirectURL,
		Scopes:       cfg.SSO.Google.Scopes,
	})
	ssoProviders[providerMicrosoft] = sso.NewMicrosoftProvider(sso.ProviderConfig{
		ClientID:     cfg.SSO.Microsoft.ClientID,
		ClientSecret: cfg.SSO.Microsoft.ClientSecret,
		RedirectURL:  cfg.SSO.Microsoft.RedirectURL,
		Scopes:       cfg.SSO.Microsoft.Scopes,
	})
	ssoProviders[providerGitHub] = sso.NewGitHubProvider(sso.ProviderConfig{
		ClientID:     cfg.SSO.GitHub.ClientID,
		ClientSecret: cfg.SSO.GitHub.ClientSecret,
		RedirectURL:  cfg.SSO.GitHub.RedirectURL,
		Scopes:       cfg.SSO.GitHub.Scopes,
	})
	return ssoProviders
}

func provideRoleRepository(sqlxDB *sqlx.DB, logger *logrus.Logger) roleRepository.RoleRepository {
	return roleRepository.NewRoleRepository(sqlxDB, logger)
}

func provideOrganizationRepository(sqlxDB *sqlx.DB, redisClient *redis.Client) orgRepo.OrganizationRepository {
	return orgRepo.NewOrganizationRepository(sqlxDB, redisClient)
}

func provideAuditModule(
	sqlxDB *sqlx.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	wsManager *ws2.WebSocketManager,
	taskDistributor worker.TaskDistributor,
) *audit.AuditModule {
	return audit.NewAuditModule(sqlxDB, logger, validate, wsManager, taskDistributor)
}

type authModuleParams struct {
	fx.In

	Cfg             *AppConfig
	JWTManager      *jwt.JWTManager
	DB              *sqlx.DB
	Redis           *redis.Client
	Logger          *logrus.Logger
	Validate        *validator.Validate
	TM              tx.WithTransactionManager
	WSManager       *ws2.WebSocketManager
	SSEManager      *sse.Manager
	Enforcer        usecase.IEnforcer
	AuditMod        *audit.AuditModule
	TaskDistributor worker.TaskDistributor
	OrgRepo         orgRepo.OrganizationRepository
	TicketManager   ws2.TicketManager
	SSOProviders    map[string]sso.Provider
}

func provideAuthModule(p authModuleParams) *auth.AuthModule {
	return auth.NewAuthModule(
		p.Cfg.Security.MaxLoginAttempts,
		p.Cfg.Security.LockoutDuration,
		p.Cfg.Security.MaxConcurrentSessions,
		p.JWTManager,
		p.DB,
		p.Redis,
		p.Logger,
		p.Validate,
		p.TM,
		p.WSManager,
		p.SSEManager,
		p.Enforcer,
		p.AuditMod,
		p.TaskDistributor,
		p.OrgRepo,
		p.TicketManager,
		p.Cfg.Casbin.DefaultRole,
		p.Cfg.Casbin.DefaultDomain,
		p.SSOProviders,
		p.Cfg.Server.FrontendBaseURL,
		authHttp.CookieConfig{
			Domain:   p.Cfg.Cookie.Domain,
			SameSite: p.Cfg.Cookie.SameSite,
			Secure:   p.Cfg.Cookie.Secure,
		},
	)
}

func provideWebhookModule(
	sqlxDB *sqlx.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	taskDistributor worker.TaskDistributor,
) *webhook.WebhookModule {
	return webhook.NewWebhookModule(sqlxDB, logger, validate, taskDistributor)
}

type userModuleParams struct {
	fx.In

	DB              *sqlx.DB
	Logger          *logrus.Logger
	Validate        *validator.Validate
	TM              tx.WithTransactionManager
	Enforcer        usecase.IEnforcer
	AuditMod        *audit.AuditModule
	AuthMod         *auth.AuthModule
	WebhookMod      *webhook.WebhookModule
	StorageProvider storage.Provider
}

func provideUserModule(p userModuleParams) *user.UserModule {
	return user.NewUserModule(
		p.DB,
		p.Logger,
		p.Validate,
		p.TM,
		p.Enforcer,
		p.AuditMod,
		p.AuthMod,
		p.WebhookMod,
		p.StorageProvider,
	)
}

func provideAPIKeyModule(
	sqlxDB *sqlx.DB,
	userMod *user.UserModule,
	redisClient *redis.Client,
	logger *logrus.Logger,
	validate *validator.Validate,
) *api_key.ApiKeyModule {
	return api_key.NewApiKeyModule(sqlxDB, userMod.UserRepo, redisClient, logger, validate)
}

func provideAccessModule(
	sqlxDB *sqlx.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
) *access.AccessModule {
	return access.NewAccessModule(sqlxDB, logger, validate)
}

func providePermissionModule(
	enforcer usecase.IEnforcer,
	validate *validator.Validate,
	logger *logrus.Logger,
	roleRepo roleRepository.RoleRepository,
	userMod *user.UserModule,
	accessMod *access.AccessModule,
	auditMod *audit.AuditModule,
) *permission.PermissionModule {
	return permission.NewPermissionModule(enforcer, validate, logger, roleRepo, userMod.UserRepo, accessMod.AccessRepo, auditMod)
}

func provideRoleModule(
	sqlxDB *sqlx.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	tm tx.WithTransactionManager,
	permissionMod *permission.PermissionModule,
) *role.RoleModule {
	return role.NewRoleModule(sqlxDB, logger, validate, tm, permissionMod.PermissionUseCase)
}

func provideStatsModule(sqlxDB *sqlx.DB, logger *logrus.Logger) *stats.StatsModule {
	return stats.NewStatsModule(sqlxDB, logger)
}

func provideProjectModule(sqlxDB *sqlx.DB, validate *validator.Validate) *project.ProjectModule {
	return project.NewProjectModule(sqlxDB, validate)
}

type organizationModuleParams struct {
	fx.In

	Cfg             *AppConfig
	DB              *sqlx.DB
	Redis           *redis.Client
	TaskDistributor worker.TaskDistributor
	UserMod         *user.UserModule
	Logger          *logrus.Logger
	Validate        *validator.Validate
	TM              tx.WithTransactionManager
	Enforcer        usecase.IEnforcer
	PresenceManager ws2.PresenceManager
	RoleRepo        roleRepository.RoleRepository
}

func provideOrganizationModule(p organizationModuleParams) *organization.OrganizationModule {
	return organization.NewOrganizationModule(
		p.DB,
		p.Redis,
		p.TaskDistributor,
		p.UserMod.UserRepo,
		p.Logger,
		p.Validate,
		p.TM,
		p.Enforcer,
		p.PresenceManager,
		p.Cfg.Server.FrontendBaseURL,
		p.RoleRepo,
	)
}

// MetricsBroadcaster manages the lifecycle of the real-time metrics broadcaster
// and user presence pruner goroutines.
type MetricsBroadcaster struct {
	wsManager       *ws2.WebSocketManager
	presenceManager ws2.PresenceManager
	statsModule     *stats.StatsModule

	mu     sync.Mutex
	cancel context.CancelFunc
}

func provideMetricsBroadcaster(
	wsManager *ws2.WebSocketManager,
	presenceManager ws2.PresenceManager,
	statsModule *stats.StatsModule,
) *MetricsBroadcaster {
	return &MetricsBroadcaster{
		wsManager:       wsManager,
		presenceManager: presenceManager,
		statsModule:     statsModule,
	}
}

func (b *MetricsBroadcaster) Start() context.CancelFunc {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cancel != nil {
		return b.cancel
	}
	ctx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel
	startMetricsBroadcaster(ctx, b.wsManager, b.presenceManager, b.statsModule)
	return cancel
}

func (b *MetricsBroadcaster) Stop() {
	b.mu.Lock()
	cancel := b.cancel
	b.cancel = nil
	b.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func startMetricsBroadcaster(
	ctx context.Context,
	wsManager *ws2.WebSocketManager,
	presenceManager ws2.PresenceManager,
	statsModule *stats.StatsModule,
) {
	// Goroutine 1: Metrics Broadcaster (2s tick)
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		var lastCount uint64
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				currentCount := middleware.GetTotalRequests()
				rps := float64(currentCount-lastCount) / 2.0
				lastCount = currentCount

				reqCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
				sysStats, err := statsModule.UseCase.GetSystemInsights(reqCtx)
				summary, _ := statsModule.UseCase.GetDashboardSummary(reqCtx)
				cancel()

				if err != nil || summary == nil {
					continue
				}

				payload, _ := json.Marshal(map[string]interface{}{
					"type": metricsEventType,
					"data": map[string]interface{}{
						"scope":            "instance",
						"rps":              rps,
						"active_users":     wsManager.ClientCount(),
						"total_users":      summary.TotalUsers,
						"most_active_role": sysStats.MostActiveRole,
					},
				})
				wsManager.BroadcastToChannel(metricsChannel, payload)
			}
		}
	}()

	// Goroutine 2: Presence Pruner (5m tick)
	go func() {
		ticker := time.NewTicker(defaultPresencePruneInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				removed, err := presenceManager.PruneStaleUsers(reqCtx, defaultPresencePruneInterval)
				cancel()

				if err == nil {
					for orgID, userIDs := range removed {
						for _, uid := range userIDs {
							wsManager.PresenceUpdate(orgID, "leave", &ws2.PresenceUser{UserID: uid})
						}
					}
				}
			}
		}
	}()
}

func provideTusHandler(
	cfg *AppConfig,
	logger *logrus.Logger,
	userModule *user.UserModule,
	redisClient *redis.Client,
) *handler.Handler {
	tusHandler, err := initTusHandler(cfg, logger, userModule, redisClient)
	if err != nil {
		logger.Errorf("Failed to init TUS handler: %v", err)
		return nil
	}
	logger.Info("TUS Handler initialized.")
	return tusHandler
}

func initTusHandler(
	cfg *AppConfig,
	logger *logrus.Logger,
	userModule *user.UserModule,
	redisClient *redis.Client,
) (*handler.Handler, error) {
	tusRegistry := tus.NewRegistry()

	tusRegistry.Register("avatar", &userUseCase.AvatarHook{
		UserUseCase: userModule.UserUseCase,
		Tracker:     userUseCase.NewRedisProcessedUploadTracker(redisClient),
	})

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

	return tus.NewHandler(tus.Config{
		StorageDriver: cfg.Storage.Driver,
		LocalRootPath: cfg.Storage.Local.RootPath,
		S3Bucket:      cfg.Storage.S3.Bucket,
		S3Endpoint:    cfg.Storage.S3.Endpoint,
		BasePath:      cfg.Tus.BasePath,
	}, tusRegistry, s3Client, logger)
}
