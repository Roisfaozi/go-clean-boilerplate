package router

import (
	"net/http"

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
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/tus/tusd/v2/pkg/handler"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type StdRouterConfig struct {
	AllowedOrigins []string
	MetricsEnabled bool
	MetricsAuth    bool
	MetricsUser    string
	MetricsPass    string
	OTEL           struct {
		Enabled     bool
		ServiceName string
	}
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
	db any,
	redisClient *redis.Client,
	tusHandler *handler.Handler,
	logger *logrus.Logger,
) http.Handler {
	mux := http.NewServeMux()

	// Base middleware layers
	reqID := middleware.HTTPRequestIDMiddleware()
	reqLog := middleware.HTTPRequestLogger(logger)
	rec := middleware.HTTPRecoveryMiddleware(logger)
	sec := middleware.HTTPSecurityMiddleware()
	cors := middleware.HTTPCORSMiddleware(cfg.AllowedOrigins)

	authMw := authMiddleware.HTTPValidateToken()
	wsTicketMw := authMiddleware.HTTPValidateWebSocketToken()
	apiKeyAuth := apiKeyMiddleware.HTTPAuthenticate()
	apiKeyAutoScope := apiKeyMiddleware.HTTPRequireScopeAuto()
	csrf := middleware.HTTPCSRFMiddleware(logger)
	userStatus := middleware.HTTPUserStatusMiddleware(userModule.UserRepo, logger)
	requireOrg := tenantMiddleware.HTTPRequireOrganization()
	optOrg := tenantMiddleware.HTTPOptionalOrganization()
	casbinMw := middleware.HTTPCasbinMiddleware(casbinEnforcer, logger)

	// Chain definitions for route strata
	publicChain := func(h http.HandlerFunc) http.Handler {
		return delivery.Chain(h, reqID, reqLog, rec, sec, cors)
	}

	authChain := func(h http.HandlerFunc) http.Handler {
		return delivery.Chain(h, reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyAutoScope, userStatus)
	}

	tenantChain := func(h http.HandlerFunc) http.Handler {
		return delivery.Chain(h, reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyAutoScope, userStatus, requireOrg, casbinMw)
	}

	adminChain := func(h http.HandlerFunc) http.Handler {
		return delivery.Chain(h, reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("admin:manage"), userStatus, optOrg, casbinMw)
	}

	// 1. Health & System
	mux.Handle("GET /api/v1/health", publicChain(GetStdHealth(db, redisClient)))

	if cfg.MetricsEnabled {
		metricsHandler := promhttp.Handler()
		if cfg.MetricsAuth {
			metricsHandler = BasicAuthHandler(metricsHandler, cfg.MetricsUser, cfg.MetricsPass)
		}
		mux.Handle("GET /metrics", publicChain(metricsHandler.ServeHTTP))
	}

	// 2. Realtime & Uploads
	mux.Handle("GET /api/v1/events", delivery.Chain(http.HandlerFunc(sseManager.ServeHTTP), reqID, reqLog, rec, sec, cors, authMw))
	// WS upgrade requires raw http.Hijacker connection.
	// Skip reqLog/sec/cors wrappers; they wrap ResponseWriter and break gorilla upgrade.
	mux.Handle("GET /api/v1/ws", delivery.Chain(http.HandlerFunc(wsController.HandleWebSocketHTTP), reqID, rec, wsTicketMw))

	if tusHandler != nil {
		mux.Handle("/api/v1/upload/files/", delivery.Chain(http.StripPrefix("/api/v1/upload/files/", tusHandler), reqID, reqLog, rec, sec, cors, authMw, userStatus))
	}

	// 3. Auth Routes
	mux.Handle("POST /api/v1/auth/register", publicChain(authModule.AuthController.HTTPRegister))
	mux.Handle("POST /api/v1/auth/login", publicChain(authModule.AuthController.HTTPLogin))
	mux.Handle("POST /api/v1/auth/refresh", publicChain(authModule.AuthController.HTTPRefreshToken))
	mux.Handle("POST /api/v1/auth/forgot-password", publicChain(authModule.AuthController.HTTPForgotPassword))
	mux.Handle("POST /api/v1/auth/reset-password", publicChain(authModule.AuthController.HTTPResetPassword))
	mux.Handle("POST /api/v1/auth/verify-email", publicChain(authModule.AuthController.HTTPVerifyEmail))
	mux.Handle("GET /api/v1/auth/sso/{provider}", publicChain(authModule.AuthController.HTTPSSOLogin))
	mux.Handle("GET /api/v1/auth/sso/{provider}/callback", publicChain(authModule.AuthController.HTTPSSOCallback))

	mux.Handle("POST /api/v1/auth/logout", authChain(authModule.AuthController.HTTPLogout))
	mux.Handle("POST /api/v1/auth/ticket", authChain(authModule.AuthController.HTTPGetTicket))
	mux.Handle("POST /api/v1/auth/resend-verification", authChain(authModule.AuthController.HTTPResendVerification))
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
	mux.Handle("POST /api/v1/organizations/invitations/accept", publicChain(organizationModule.OrganizationController.HTTPAcceptInvitation))
	mux.Handle("POST /api/v1/organizations", authChain(organizationModule.OrganizationController.HTTPCreateOrganization))
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
	mux.Handle("DELETE /api/v1/organizations/{id}", delivery.Chain(http.HandlerFunc(organizationModule.OrganizationController.HTTPDeleteOrganization), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireUserSession(), userStatus, requireOrg, casbinMw))

	mux.Handle("POST /api/v1/organizations/{id}/members/invite", delivery.Chain(http.HandlerFunc(organizationModule.OrganizationController.HTTPInviteMember), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("member:manage"), userStatus, requireOrg, casbinMw))
	mux.Handle("PATCH /api/v1/organizations/{id}/members/{userId}", delivery.Chain(http.HandlerFunc(organizationModule.OrganizationController.HTTPUpdateMemberRole), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("member:manage"), userStatus, requireOrg, casbinMw))
	mux.Handle("DELETE /api/v1/organizations/{id}/members/{userId}", delivery.Chain(http.HandlerFunc(organizationModule.OrganizationController.HTTPRemoveMember), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("member:manage"), userStatus, requireOrg, casbinMw))

	mux.Handle("POST /api/v1/organizations/{id}/restore", delivery.Chain(http.HandlerFunc(organizationModule.OrganizationController.HTTPRestoreOrganization), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireUserSession(), userStatus, optOrg, casbinMw))
	mux.Handle("DELETE /api/v1/organizations/{id}/hard", delivery.Chain(http.HandlerFunc(organizationModule.OrganizationController.HTTPHardDeleteOrganization), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireUserSession(), userStatus, optOrg, casbinMw))

	// 6. Projects Routes
	mux.Handle("POST /api/v1/projects", delivery.Chain(http.HandlerFunc(projectModule.ProjectController.HTTPCreate), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("project:manage"), userStatus, requireOrg, casbinMw))
	mux.Handle("GET /api/v1/projects", delivery.Chain(http.HandlerFunc(projectModule.ProjectController.HTTPGetAll), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("project:view", "project:manage"), userStatus, requireOrg, casbinMw))
	mux.Handle("GET /api/v1/projects/{id}", delivery.Chain(http.HandlerFunc(projectModule.ProjectController.HTTPGetByID), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("project:view", "project:manage"), userStatus, requireOrg, casbinMw))
	mux.Handle("PUT /api/v1/projects/{id}", delivery.Chain(http.HandlerFunc(projectModule.ProjectController.HTTPUpdate), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("project:manage"), userStatus, requireOrg, casbinMw))
	mux.Handle("DELETE /api/v1/projects/{id}", delivery.Chain(http.HandlerFunc(projectModule.ProjectController.HTTPDelete), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("project:manage"), userStatus, requireOrg, casbinMw))

	// 7. Roles & Organization Roles
	mux.Handle("POST /api/v1/organizations/{id}/roles", delivery.Chain(http.HandlerFunc(roleModule.RoleController.HTTPCreateOrganizationRole), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("role:manage"), userStatus, requireOrg, casbinMw))
	mux.Handle("PUT /api/v1/organizations/{id}/roles/{roleId}", delivery.Chain(http.HandlerFunc(roleModule.RoleController.HTTPUpdateOrganizationRole), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("role:manage"), userStatus, requireOrg, casbinMw))
	mux.Handle("DELETE /api/v1/organizations/{id}/roles/{roleId}", delivery.Chain(http.HandlerFunc(roleModule.RoleController.HTTPDeleteOrganizationRole), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("role:manage"), userStatus, requireOrg, casbinMw))

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
	mux.Handle("POST /api/v1/access-rights", adminChain(accessModule.AccessController.HTTPCreateAccessRight))
	mux.Handle("GET /api/v1/access-rights", adminChain(accessModule.AccessController.HTTPGetAllAccessRights))
	mux.Handle("POST /api/v1/access-rights/search", adminChain(accessModule.AccessController.HTTPGetAccessRightsDynamic))
	mux.Handle("DELETE /api/v1/access-rights/{id}", adminChain(accessModule.AccessController.HTTPDeleteAccessRight))
	mux.Handle("POST /api/v1/access-rights/link", adminChain(accessModule.AccessController.HTTPLinkEndpointToAccessRight))
	mux.Handle("POST /api/v1/access-rights/unlink", adminChain(accessModule.AccessController.HTTPUnlinkEndpointFromAccessRight))
	mux.Handle("POST /api/v1/endpoints", adminChain(accessModule.AccessController.HTTPCreateEndpoint))
	mux.Handle("POST /api/v1/endpoints/search", adminChain(accessModule.AccessController.HTTPGetEndpointsDynamic))
	mux.Handle("DELETE /api/v1/endpoints/{id}", adminChain(accessModule.AccessController.HTTPDeleteEndpoint))

	// 10. API Keys Routes
	mux.Handle("POST /api/v1/api-keys", tenantChain(apiKeyModule.Controller.HTTPCreate))
	mux.Handle("GET /api/v1/api-keys", tenantChain(apiKeyModule.Controller.HTTPList))
	mux.Handle("DELETE /api/v1/api-keys/{id}", tenantChain(apiKeyModule.Controller.HTTPRevoke))

	// 11. Webhooks Routes
	mux.Handle("POST /api/v1/webhooks", delivery.Chain(http.HandlerFunc(webhookModule.Controller.HTTPCreate), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("webhook:manage"), userStatus, requireOrg, casbinMw))
	mux.Handle("GET /api/v1/webhooks", delivery.Chain(http.HandlerFunc(webhookModule.Controller.HTTPFindByOrganization), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("webhook:manage"), userStatus, requireOrg, casbinMw))
	mux.Handle("GET /api/v1/webhooks/{id}", delivery.Chain(http.HandlerFunc(webhookModule.Controller.HTTPFindByID), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("webhook:manage"), userStatus, requireOrg, casbinMw))
	mux.Handle("PUT /api/v1/webhooks/{id}", delivery.Chain(http.HandlerFunc(webhookModule.Controller.HTTPUpdate), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("webhook:manage"), userStatus, requireOrg, casbinMw))
	mux.Handle("DELETE /api/v1/webhooks/{id}", delivery.Chain(http.HandlerFunc(webhookModule.Controller.HTTPDelete), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("webhook:manage"), userStatus, requireOrg, casbinMw))
	mux.Handle("GET /api/v1/webhooks/{id}/logs", delivery.Chain(http.HandlerFunc(webhookModule.Controller.HTTPGetLogs), reqID, reqLog, rec, sec, cors, apiKeyAuth, authMw, csrf, apiKeyMiddleware.HTTPRequireScopes("webhook:manage"), userStatus, requireOrg, casbinMw))

	// 12. Audit Logs Routes
	mux.Handle("POST /api/v1/audit-logs/search", adminChain(auditModule.AuditController.HTTPGetLogsDynamic))
	mux.Handle("GET /api/v1/audit-logs/export", adminChain(auditModule.AuditController.HTTPExport))
	mux.Handle("GET /api/v1/audit-logs/export-async", adminChain(auditModule.AuditController.HTTPExportAsync))

	// 13. Stats Routes
	mux.Handle("GET /api/v1/stats/summary", authChain(statsModule.StatsController.HTTPSummary))
	mux.Handle("GET /api/v1/stats/activity", authChain(statsModule.StatsController.HTTPActivity))
	mux.Handle("GET /api/v1/stats/insights", authChain(statsModule.StatsController.HTTPInsights))

	var handler http.Handler = mux
	if cfg.OTEL.Enabled {
		handler = otelhttp.NewHandler(mux, cfg.OTEL.ServiceName)
	}

	return handler
}

// GetStdHealth returns http.HandlerFunc checking MySQL and Redis connectivity.
func GetStdHealth(db any, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "OK"
		details := make(map[string]string)

		if db != nil {
			if sqlxDB := tx.ExtractSQLX(db); sqlxDB != nil {
				if err := sqlxDB.PingContext(r.Context()); err != nil {
					status = "DEGRADED"
					details["mysql"] = "DOWN"
				} else {
					details["mysql"] = "UP"
				}
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
