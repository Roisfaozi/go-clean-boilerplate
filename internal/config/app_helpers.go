package config

import (
	"context"
	"encoding/json"
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
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/tus/tusd/v2/pkg/handler"
	"gorm.io/gorm"
)

type appModules struct {
	audit        *audit.AuditModule
	auth         *auth.AuthModule
	webhook      *webhook.WebhookModule
	user         *user.UserModule
	apiKey       *api_key.ApiKeyModule
	access       *access.AccessModule
	permission   *permission.PermissionModule
	role         *role.RoleModule
	stats        *stats.StatsModule
	project      *project.ProjectModule
	organization *organization.OrganizationModule
}

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

func initModules(
	cfg *AppConfig,
	dbConnection *gorm.DB,
	redisClient *redis.Client,
	logger *logrus.Logger,
	validate *validator.Validate,
	wsManager *ws2.WebSocketManager,
	presenceManager ws2.PresenceManager,
	taskDistributor worker.TaskDistributor,
	tm tx.WithTransactionManager,
	jwtManager *jwt.JWTManager,
	ticketManager ws2.TicketManager,
	enforcer usecase.IEnforcer,
	sseManager *sse.Manager,
	storageProvider storage.Provider,
	ssoProviders map[string]sso.Provider,
) appModules {
	roleRepo := roleRepository.NewRoleRepository(dbConnection, logger)
	organizationRepository := orgRepo.NewOrganizationRepository(dbConnection, redisClient)

	auditModule := audit.NewAuditModule(dbConnection, logger, validate, wsManager, taskDistributor)

	authModule := auth.NewAuthModule(
		cfg.Security.MaxLoginAttempts,
		cfg.Security.LockoutDuration,
		cfg.Security.MaxConcurrentSessions,
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
		ssoProviders,
		cfg.Server.FrontendBaseURL,
		authHttp.CookieConfig{
			Domain:   cfg.Cookie.Domain,
			SameSite: cfg.Cookie.SameSite,
			Secure:   cfg.Cookie.Secure,
		},
	)

	webhookModule := webhook.NewWebhookModule(dbConnection, logger, validate, taskDistributor)

	userModule := user.NewUserModule(dbConnection, logger, validate, tm, enforcer, auditModule, authModule, webhookModule, storageProvider)

	apiKeyModule := api_key.NewApiKeyModule(dbConnection, userModule.UserRepo, redisClient, logger, validate)

	accessModule := access.NewAccessModule(dbConnection, logger, validate)

	permissionModule := permission.NewPermissionModule(enforcer, validate, logger, roleRepo, userModule.UserRepo, accessModule.AccessRepo, auditModule)

	roleModule := role.NewRoleModule(dbConnection, logger, validate, tm, permissionModule.PermissionUseCase)

	statsModule := stats.NewStatsModule(dbConnection, logger)

	projectModule := project.NewProjectModule(dbConnection, validate)

	organizationModule := organization.NewOrganizationModule(dbConnection, redisClient, taskDistributor, userModule.UserRepo, logger, validate, tm, enforcer, presenceManager, cfg.Server.FrontendBaseURL, roleRepo)

	return appModules{
		audit:        auditModule,
		auth:         authModule,
		webhook:      webhookModule,
		user:         userModule,
		apiKey:       apiKeyModule,
		access:       accessModule,
		permission:   permissionModule,
		role:         roleModule,
		stats:        statsModule,
		project:      projectModule,
		organization: organizationModule,
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

func initTusHandler(
	cfg *AppConfig,
	logger *logrus.Logger,
	userModule *user.UserModule,
) (*handler.Handler, error) {
	tusRegistry := tus.NewRegistry()

	tusRegistry.Register("avatar", &userUseCase.AvatarHook{UserUseCase: userModule.UserUseCase})

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
