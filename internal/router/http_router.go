package router

import (
	"net/http"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/tus/tusd/v2/pkg/handler"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type StdRouterConfig struct {
	AllowedOrigins   []string
	TrustedProxies   []string
	RateLimitEnabled bool
	RateLimitRPS     float64
	RateLimitBurst   int
	RateLimitStore   string
	MetricsEnabled   bool
	MetricsAuth      bool
	MetricsUser      string
	MetricsPass      string
	OTEL             struct {
		Enabled     bool
		ServiceName string
	}
}

func chainNonNil(h http.Handler, middlewares ...delivery.Middleware) http.Handler {
	filtered := make([]delivery.Middleware, 0, len(middlewares))
	for _, mw := range middlewares {
		if mw != nil {
			filtered = append(filtered, mw)
		}
	}
	return delivery.Chain(h, filtered...)
}

// SetupStdRouter wires all route handlers on a Go 1.25 standard http.ServeMux.
func SetupStdRouter(
	cfg StdRouterConfig,
	authModule *auth.AuthModule,
	userModule *user.UserModule,
	permissionModule *permission.PermissionModule,
	accessModule *access.AccessModule,
	roleModule *role.RoleModule,
	organizationModule *organization.OrganizationModule,
	auditModule *audit.AuditModule,
	statsModule *stats.StatsModule,
	projectModule *project.ProjectModule,
	apiKeyModule *api_key.ApiKeyModule,
	webhookModule *webhook.WebhookModule,
	authMiddleware *middleware.AuthMiddleware,
	apiKeyMiddleware *middleware.APIKeyMiddleware,
	casbinEnforcer middleware.CasbinEnforcer,
	tenantMiddleware *middleware.TenantMiddleware,
	wsController *ws.WebSocketController,
	sseManager *sse.Manager,
	db *sqlx.DB,
	redisClient *redis.Client,
	tusHandler *handler.Handler,
	logger *logrus.Logger,
) http.Handler {
	mux := http.NewServeMux()

	// Base middleware layers
	reqID := middleware.HTTPRequestIDMiddleware()
	var promMw delivery.Middleware
	if cfg.MetricsEnabled {
		promMw = middleware.HTTPPrometheusMiddleware()
	}
	reqLog := middleware.HTTPRequestLogger(logger)
	rec := middleware.HTTPRecoveryMiddleware(logger)
	sec := middleware.HTTPSecurityMiddleware()
	cors := middleware.HTTPCORSMiddleware(cfg.AllowedOrigins)
	trustedProxy := middleware.HTTPTrustedProxyMiddleware(cfg.TrustedProxies, logger)

	// 3-tier rate limiters (Redis) or global memory fallback
	var publicLimiter, criticalLimiter, authLimiter, globalMemoryLimiter delivery.Middleware
	if cfg.RateLimitEnabled {
		if cfg.RateLimitStore == "redis" {
			// Tier 1: Public API - 10 RPS (600 RPM)
			publicLimiter = middleware.HTTPRateLimitMiddlewareRedis(redisClient, logger, middleware.LimiterTypeIP, 10*60, 60)
			// Tier 3: Critical Endpoints (Login) - 5 RPM
			criticalLimiter = middleware.HTTPRateLimitMiddlewareRedis(redisClient, logger, middleware.LimiterTypeIP, 5, 60)
			// Tier 2: Authenticated User - 100 RPS (6000 RPM)
			authLimiter = middleware.HTTPRateLimitMiddlewareRedis(redisClient, logger, middleware.LimiterTypeUser, 100*60, 60)
			if logger != nil {
				logger.Info("Advanced Rate Limiter enabled: Redis store")
			}
		} else {
			rps := cfg.RateLimitRPS
			if rps <= 0 {
				rps = 10.0
			}
			burst := cfg.RateLimitBurst
			if burst <= 0 {
				burst = 20
			}
			globalMemoryLimiter = middleware.HTTPRateLimitMiddlewareMemory(rps, burst)
			if logger != nil {
				logger.Info("Rate Limiter enabled: Memory store")
			}
		}
	}

	idempotencyMw := middleware.HTTPIdempotencyMiddleware(redisClient, logger, 24*time.Hour)

	authMw := authMiddleware.HTTPValidateToken()
	wsTicketMw := authMiddleware.HTTPValidateWebSocketToken()
	apiKeyAuth := apiKeyMiddleware.HTTPAuthenticate()
	apiKeyAutoScope := apiKeyMiddleware.HTTPRequireScopeAuto()
	requireUserSession := apiKeyMiddleware.HTTPRequireUserSession()
	adminScope := apiKeyMiddleware.HTTPRequireScopes("admin:manage")
	csrf := middleware.HTTPCSRFMiddleware(logger)
	userStatus := middleware.HTTPUserStatusMiddleware(userModule.UserRepo, logger)
	requireOrg := tenantMiddleware.HTTPRequireOrganization()
	optOrg := tenantMiddleware.HTTPOptionalOrganization()
	casbinMw := middleware.HTTPCasbinMiddleware(casbinEnforcer, logger)

	// Chain definitions for route strata
	systemChain := func(h http.HandlerFunc, extra ...delivery.Middleware) http.Handler {
		mws := []delivery.Middleware{reqID, promMw, reqLog, rec, sec}
		mws = append(mws, extra...)
		return chainNonNil(h, mws...)
	}

	publicChain := func(h http.HandlerFunc, extra ...delivery.Middleware) http.Handler {
		mws := []delivery.Middleware{reqID, promMw, reqLog, rec, sec, publicLimiter}
		mws = append(mws, extra...)
		return chainNonNil(h, mws...)
	}

	authChain := func(h http.HandlerFunc, extra ...delivery.Middleware) http.Handler {
		mws := []delivery.Middleware{reqID, promMw, reqLog, rec, sec, apiKeyAuth, authMw, csrf, apiKeyAutoScope, requireUserSession, userStatus, authLimiter}
		mws = append(mws, extra...)
		return chainNonNil(h, mws...)
	}

	tenantChain := func(h http.HandlerFunc, extra ...delivery.Middleware) http.Handler {
		mws := []delivery.Middleware{reqID, promMw, reqLog, rec, sec, apiKeyAuth, authMw, csrf, apiKeyAutoScope, userStatus, requireOrg, casbinMw, authLimiter}
		mws = append(mws, extra...)
		return chainNonNil(h, mws...)
	}

	tenantScopedChain := func(h http.HandlerFunc, scopeMw delivery.Middleware, extra ...delivery.Middleware) http.Handler {
		mws := []delivery.Middleware{reqID, promMw, reqLog, rec, sec, apiKeyAuth, authMw, csrf, scopeMw, userStatus, requireOrg, casbinMw, authLimiter}
		mws = append(mws, extra...)
		return chainNonNil(h, mws...)
	}

	adminChain := func(h http.HandlerFunc, extra ...delivery.Middleware) http.Handler {
		mws := []delivery.Middleware{reqID, promMw, reqLog, rec, sec, apiKeyAuth, authMw, csrf, adminScope, userStatus, optOrg, casbinMw, authLimiter}
		mws = append(mws, extra...)
		return chainNonNil(h, mws...)
	}

	adminSessionChain := func(h http.HandlerFunc, extra ...delivery.Middleware) http.Handler {
		mws := []delivery.Middleware{reqID, promMw, reqLog, rec, sec, apiKeyAuth, authMw, csrf, requireUserSession, userStatus, optOrg, casbinMw, authLimiter}
		mws = append(mws, extra...)
		return chainNonNil(h, mws...)
	}

	apiKeyChain := func(h http.HandlerFunc, extra ...delivery.Middleware) http.Handler {
		mws := []delivery.Middleware{reqID, promMw, reqLog, rec, sec, apiKeyAuth, authMw, csrf, apiKeyAutoScope, requireUserSession, userStatus, requireOrg, authLimiter}
		mws = append(mws, extra...)
		return chainNonNil(h, mws...)
	}

	// 1. Health, Docs & System
	mux.Handle("GET /api/v1/health", systemChain(GetStdHealth(db, redisClient)))

	swaggerHandler := SwaggerHandler()
	mux.Handle("GET /api/v1/docs", systemChain(swaggerHandler))
	mux.Handle("GET /api/v1/docs/{$}", systemChain(swaggerHandler))
	mux.Handle("GET /api/v1/docs/{path...}", systemChain(swaggerHandler))

	if cfg.MetricsEnabled {
		metricsHandler := promhttp.Handler()
		if cfg.MetricsAuth {
			metricsHandler = BasicAuthHandler(metricsHandler, cfg.MetricsUser, cfg.MetricsPass)
		}
		mux.Handle("GET /metrics", systemChain(metricsHandler.ServeHTTP))
	}

	// 2. Realtime & Uploads
	mux.Handle("GET /api/v1/events", chainNonNil(http.HandlerFunc(sseManager.ServeHTTP), reqID, promMw, reqLog, rec, sec, authMw))
	// WS upgrade requires raw http.Hijacker connection.
	// Skip reqLog/sec wrappers; prometheusResponseWriter and recoveryResponseWriter preserve Hijack.
	mux.Handle("GET /api/v1/ws", chainNonNil(http.HandlerFunc(wsController.HandleWebSocketHTTP), reqID, promMw, rec, wsTicketMw))

	if tusHandler != nil {
		mux.Handle("/api/v1/upload/files/", chainNonNil(http.StripPrefix("/api/v1/upload/files/", tusHandler), reqID, promMw, reqLog, rec, sec, authMw, userStatus))
	}

	// 3. Auth Routes
	mux.Handle("POST /api/v1/auth/register", publicChain(authModule.AuthController.HTTPRegister, idempotencyMw))
	mux.Handle("POST /api/v1/auth/login", publicChain(authModule.AuthController.HTTPLogin, criticalLimiter))
	mux.Handle("POST /api/v1/auth/refresh", publicChain(authModule.AuthController.HTTPRefreshToken, csrf, idempotencyMw))
	mux.Handle("POST /api/v1/auth/forgot-password", publicChain(authModule.AuthController.HTTPForgotPassword, idempotencyMw))
	mux.Handle("POST /api/v1/auth/reset-password", publicChain(authModule.AuthController.HTTPResetPassword, idempotencyMw))
	mux.Handle("POST /api/v1/auth/verify-email", publicChain(authModule.AuthController.HTTPVerifyEmail, idempotencyMw))
	mux.Handle("GET /api/v1/auth/sso/{provider}", publicChain(authModule.AuthController.HTTPSSOLogin))
	mux.Handle("GET /api/v1/auth/sso/{provider}/callback", publicChain(authModule.AuthController.HTTPSSOCallback))

	mux.Handle("POST /api/v1/auth/logout", authChain(authModule.AuthController.HTTPLogout))
	mux.Handle("POST /api/v1/auth/ticket", authChain(authModule.AuthController.HTTPGetTicket, idempotencyMw))
	mux.Handle("POST /api/v1/auth/resend-verification", authChain(authModule.AuthController.HTTPResendVerification, idempotencyMw))
	mux.Handle("GET /api/v1/auth/me", authChain(authModule.AuthController.HTTPMe))

	// 4. Users Routes
	mux.Handle("POST /api/v1/users/register", publicChain(userModule.UserController.HTTPRegisterUser))
	mux.Handle("GET /api/v1/users/me", authChain(userModule.UserController.HTTPGetCurrentUser))
	mux.Handle("PUT /api/v1/users/me", authChain(userModule.UserController.HTTPUpdateUser))
	mux.Handle("PATCH /api/v1/users/me/avatar", authChain(userModule.UserController.HTTPUpdateAvatar))

	mux.Handle("GET /api/v1/users", adminChain(userModule.UserController.HTTPGetAllUsers))
	mux.Handle("POST /api/v1/users/search", adminChain(userModule.UserController.HTTPGetUsersDynamic))
	mux.Handle("GET /api/v1/users/{id}", adminChain(userModule.UserController.HTTPGetUserByID))
	mux.Handle("PATCH /api/v1/users/{id}/status", adminChain(userModule.UserController.HTTPUpdateUserStatus))
	mux.Handle("DELETE /api/v1/users/{id}", adminChain(userModule.UserController.HTTPDeleteUser))

	// 5. Organizations Routes
	mux.Handle("POST /api/v1/organizations/invitations/accept", publicChain(organizationModule.OrganizationController.HTTPAcceptInvitation, idempotencyMw))
	mux.Handle("POST /api/v1/organizations", authChain(organizationModule.OrganizationController.HTTPCreateOrganization, idempotencyMw))
	mux.Handle("GET /api/v1/organizations/me", authChain(organizationModule.OrganizationController.HTTPGetMyOrganizations))

	mux.Handle("GET /api/v1/organizations/{id}", tenantChain(organizationModule.OrganizationController.HTTPGetOrganization))
	// Single dispatcher for 5-segment organization GETs avoids Go ServeMux
	// pattern conflicts between "slug/{slug}" and "{id}/members|presence|roles".
	mux.Handle("GET /api/v1/organizations/{id}/{action}", tenantChain(func(w http.ResponseWriter, r *http.Request) {
		action := r.PathValue("action")
		ctxOrgID := database.GetOrganizationID(r.Context())
		if ctxOrgID == "" {
			id := r.PathValue("id")
			if id != "" {
				ctx := database.SetOrganizationContext(r.Context(), id)
				ctx = delivery.SetContextValue(ctx, delivery.OrganizationIDKey, id)
				r = r.WithContext(ctx)
			}
		}
		id := r.PathValue("id")
		if id == "slug" {
			r.SetPathValue("slug", action)
			organizationModule.OrganizationController.HTTPGetOrganizationBySlug(w, r)
			return
		}
		switch action {
		case "members":
			delivery.Chain(http.HandlerFunc(organizationModule.OrganizationController.HTTPGetMembers), apiKeyMiddleware.HTTPRequireScopes("member:manage")).ServeHTTP(w, r)
		case "presence":
			delivery.Chain(http.HandlerFunc(organizationModule.OrganizationController.HTTPGetPresence), apiKeyMiddleware.HTTPRequireScopes("presence:view")).ServeHTTP(w, r)
		case "roles":
			delivery.Chain(http.HandlerFunc(roleModule.RoleController.HTTPGetOrganizationRoles), apiKeyMiddleware.HTTPRequireScopes("role:view", "role:manage")).ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	mux.Handle("PUT /api/v1/organizations/{id}", tenantChain(organizationModule.OrganizationController.HTTPUpdateOrganization))
	mux.Handle("DELETE /api/v1/organizations/{id}", tenantScopedChain(organizationModule.OrganizationController.HTTPDeleteOrganization, requireUserSession))

	mux.Handle("POST /api/v1/organizations/{id}/members/invite", tenantScopedChain(organizationModule.OrganizationController.HTTPInviteMember, apiKeyMiddleware.HTTPRequireScopes("member:manage"), idempotencyMw))
	mux.Handle("PATCH /api/v1/organizations/{id}/members/{userId}", tenantScopedChain(organizationModule.OrganizationController.HTTPUpdateMemberRole, apiKeyMiddleware.HTTPRequireScopes("member:manage")))
	mux.Handle("DELETE /api/v1/organizations/{id}/members/{userId}", tenantScopedChain(organizationModule.OrganizationController.HTTPRemoveMember, apiKeyMiddleware.HTTPRequireScopes("member:manage")))

	mux.Handle("POST /api/v1/organizations/{id}/restore", adminSessionChain(organizationModule.OrganizationController.HTTPRestoreOrganization))
	mux.Handle("DELETE /api/v1/organizations/{id}/hard", adminSessionChain(organizationModule.OrganizationController.HTTPHardDeleteOrganization))

	// 6. Projects Routes
	mux.Handle("POST /api/v1/projects", tenantScopedChain(projectModule.ProjectController.HTTPCreate, apiKeyMiddleware.HTTPRequireScopes("project:manage"), idempotencyMw))
	mux.Handle("GET /api/v1/projects", tenantScopedChain(projectModule.ProjectController.HTTPGetAll, apiKeyMiddleware.HTTPRequireScopes("project:view", "project:manage")))
	mux.Handle("GET /api/v1/projects/{id}", tenantScopedChain(projectModule.ProjectController.HTTPGetByID, apiKeyMiddleware.HTTPRequireScopes("project:view", "project:manage")))
	mux.Handle("PUT /api/v1/projects/{id}", tenantScopedChain(projectModule.ProjectController.HTTPUpdate, apiKeyMiddleware.HTTPRequireScopes("project:manage"), idempotencyMw))
	mux.Handle("DELETE /api/v1/projects/{id}", tenantScopedChain(projectModule.ProjectController.HTTPDelete, apiKeyMiddleware.HTTPRequireScopes("project:manage")))

	// 7. Roles & Organization Roles
	mux.Handle("POST /api/v1/organizations/{id}/roles", tenantScopedChain(roleModule.RoleController.HTTPCreateOrganizationRole, apiKeyMiddleware.HTTPRequireScopes("role:manage")))
	mux.Handle("PUT /api/v1/organizations/{id}/roles/{roleId}", tenantScopedChain(roleModule.RoleController.HTTPUpdateOrganizationRole, apiKeyMiddleware.HTTPRequireScopes("role:manage")))
	mux.Handle("DELETE /api/v1/organizations/{id}/roles/{roleId}", tenantScopedChain(roleModule.RoleController.HTTPDeleteOrganizationRole, apiKeyMiddleware.HTTPRequireScopes("role:manage")))

	mux.Handle("POST /api/v1/roles", adminChain(roleModule.RoleController.HTTPCreate))
	mux.Handle("GET /api/v1/roles", adminChain(roleModule.RoleController.HTTPGetAll))
	mux.Handle("PUT /api/v1/roles/{id}", adminChain(roleModule.RoleController.HTTPUpdate))
	mux.Handle("POST /api/v1/roles/search", adminChain(roleModule.RoleController.HTTPGetRolesDynamic))
	mux.Handle("DELETE /api/v1/roles/{id}", adminChain(roleModule.RoleController.HTTPDelete))

	// 8. Permissions Routes
	mux.Handle("POST /api/v1/permissions/check-batch", authChain(permissionModule.PermissionController.HTTPBatchCheck))
	mux.Handle("POST /api/v1/permissions/assign-role", adminChain(permissionModule.PermissionController.HTTPAssignRole))
	mux.Handle("DELETE /api/v1/permissions/revoke-role", adminChain(permissionModule.PermissionController.HTTPRevokeRole))
	mux.Handle("POST /api/v1/permissions/grant", adminChain(permissionModule.PermissionController.HTTPGrantPermission))
	mux.Handle("GET /api/v1/permissions", adminChain(permissionModule.PermissionController.HTTPGetAllPermissions))
	// Single dispatcher for 4-segment permission GETs avoids Go ServeMux
	// pattern conflicts between "{role}", "resources", and "inheritance-tree".
	mux.Handle("GET /api/v1/permissions/{name}", adminChain(func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		switch name {
		case "resources":
			permissionModule.PermissionController.HTTPGetResourceAggregation(w, r)
		case "inheritance-tree":
			permissionModule.PermissionController.HTTPGetInheritanceTree(w, r)
		default:
			r.SetPathValue("role", name)
			permissionModule.PermissionController.HTTPGetPermissionsForRole(w, r)
		}
	}))
	mux.Handle("GET /api/v1/permissions/roles/{role}/users", adminChain(permissionModule.PermissionController.HTTPGetUsersForRole))
	mux.Handle("PUT /api/v1/permissions", adminChain(permissionModule.PermissionController.HTTPUpdatePermission))
	mux.Handle("DELETE /api/v1/permissions/revoke", adminChain(permissionModule.PermissionController.HTTPRevokePermission))
	mux.Handle("POST /api/v1/permissions/inheritance", adminChain(permissionModule.PermissionController.HTTPAddRoleInheritance))
	mux.Handle("DELETE /api/v1/permissions/inheritance", adminChain(permissionModule.PermissionController.HTTPRemoveRoleInheritance))
	mux.Handle("GET /api/v1/permissions/{role}/parents", adminChain(permissionModule.PermissionController.HTTPGetParentRoles))
	mux.Handle("GET /api/v1/permissions/roles/{role}/access-rights", adminChain(permissionModule.PermissionController.HTTPGetRoleAccessRights))
	mux.Handle("POST /api/v1/permissions/assign-access-right", adminChain(permissionModule.PermissionController.HTTPAssignAccessRight))
	mux.Handle("DELETE /api/v1/permissions/revoke-access-right", adminChain(permissionModule.PermissionController.HTTPRevokeAccessRight))

	// 9. Access Rights & Endpoints
	mux.Handle("POST /api/v1/access-rights", adminChain(accessModule.AccessController.HTTPCreateAccessRight, idempotencyMw))
	mux.Handle("GET /api/v1/access-rights", adminChain(accessModule.AccessController.HTTPGetAllAccessRights))
	mux.Handle("POST /api/v1/access-rights/search", adminChain(accessModule.AccessController.HTTPGetAccessRightsDynamic))
	mux.Handle("DELETE /api/v1/access-rights/{id}", adminChain(accessModule.AccessController.HTTPDeleteAccessRight))
	mux.Handle("POST /api/v1/access-rights/link", adminChain(accessModule.AccessController.HTTPLinkEndpointToAccessRight))
	mux.Handle("POST /api/v1/access-rights/unlink", adminChain(accessModule.AccessController.HTTPUnlinkEndpointFromAccessRight))
	mux.Handle("POST /api/v1/endpoints", adminChain(accessModule.AccessController.HTTPCreateEndpoint, idempotencyMw))
	mux.Handle("POST /api/v1/endpoints/search", adminChain(accessModule.AccessController.HTTPGetEndpointsDynamic))
	mux.Handle("DELETE /api/v1/endpoints/{id}", adminChain(accessModule.AccessController.HTTPDeleteEndpoint))

	// 10. API Keys Routes
	mux.Handle("POST /api/v1/api-keys", apiKeyChain(apiKeyModule.Controller.HTTPCreate, idempotencyMw))
	mux.Handle("GET /api/v1/api-keys", tenantChain(apiKeyModule.Controller.HTTPList))
	mux.Handle("DELETE /api/v1/api-keys/{id}", tenantChain(apiKeyModule.Controller.HTTPRevoke))

	// 11. Webhooks Routes
	mux.Handle("POST /api/v1/webhooks", tenantScopedChain(webhookModule.Controller.HTTPCreate, apiKeyMiddleware.HTTPRequireScopes("webhook:manage"), idempotencyMw))
	mux.Handle("GET /api/v1/webhooks", tenantScopedChain(webhookModule.Controller.HTTPFindByOrganization, apiKeyMiddleware.HTTPRequireScopes("webhook:manage")))
	mux.Handle("GET /api/v1/webhooks/{id}", tenantScopedChain(webhookModule.Controller.HTTPFindByID, apiKeyMiddleware.HTTPRequireScopes("webhook:manage")))
	mux.Handle("PUT /api/v1/webhooks/{id}", tenantScopedChain(webhookModule.Controller.HTTPUpdate, apiKeyMiddleware.HTTPRequireScopes("webhook:manage"), idempotencyMw))
	mux.Handle("DELETE /api/v1/webhooks/{id}", tenantScopedChain(webhookModule.Controller.HTTPDelete, apiKeyMiddleware.HTTPRequireScopes("webhook:manage")))
	mux.Handle("GET /api/v1/webhooks/{id}/logs", tenantScopedChain(webhookModule.Controller.HTTPGetLogs, apiKeyMiddleware.HTTPRequireScopes("webhook:manage")))

	// 12. Audit Logs Routes
	mux.Handle("POST /api/v1/audit-logs/search", adminChain(auditModule.AuditController.HTTPGetLogsDynamic))
	mux.Handle("GET /api/v1/audit-logs/export", adminChain(auditModule.AuditController.HTTPExport))
	mux.Handle("GET /api/v1/audit-logs/export-async", adminChain(auditModule.AuditController.HTTPExportAsync))

	// 13. Stats Routes
	mux.Handle("GET /api/v1/stats/summary", authChain(statsModule.StatsController.HTTPSummary))
	mux.Handle("GET /api/v1/stats/activity", authChain(statsModule.StatsController.HTTPActivity))
	mux.Handle("GET /api/v1/stats/insights", authChain(statsModule.StatsController.HTTPInsights))

	var rootHandler http.Handler = mux
	if globalMemoryLimiter != nil {
		rootHandler = globalMemoryLimiter(rootHandler)
	}
	rootHandler = cors(rootHandler)
	if len(cfg.TrustedProxies) > 0 {
		rootHandler = trustedProxy(rootHandler)
	}
	if cfg.OTEL.Enabled {
		rootHandler = otelhttp.NewHandler(rootHandler, cfg.OTEL.ServiceName)
	}

	return rootHandler
}

// GetStdHealth returns http.HandlerFunc checking MySQL and Redis connectivity.
func GetStdHealth(db *sqlx.DB, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "OK"
		details := make(map[string]string)

		if db != nil {
			if err := db.PingContext(r.Context()); err != nil {
				status = "DEGRADED"
				details["mysql"] = "DOWN"
			} else {
				details["mysql"] = "UP"
			}
		}

		if redisClient != nil {
			if err := redisClient.Ping(r.Context()).Err(); err != nil {
				status = "DEGRADED"
				details["redis"] = "DOWN"
			} else {
				details["redis"] = "UP"
			}
		}

		response.WriteJSON(w, http.StatusOK, map[string]any{
			"status":  status,
			"details": details,
		})
	}
}

// BasicAuthHandler wraps an http.Handler with HTTP Basic Auth.
func BasicAuthHandler(next http.Handler, expectedUser, expectedPass string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != expectedUser || pass != expectedPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
