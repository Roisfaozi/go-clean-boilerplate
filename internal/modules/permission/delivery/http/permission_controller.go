package http

import (
	"context"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

func respondPermissionNoopHTTP(w http.ResponseWriter, message string) {
	response.WriteSuccess(w, http.StatusOK, map[string]any{"changed": false, "message": message})
}

type PermissionController struct {
	useCase  usecase.IPermissionUseCase
	log      *logrus.Logger
	validate *validator.Validate
}

func NewPermissionController(useCase usecase.IPermissionUseCase, log *logrus.Logger, validate *validator.Validate) *PermissionController {
	return &PermissionController{
		useCase:  useCase,
		log:      log,
		validate: validate,
	}
}

// resolveDomainHTTP ensures that a tenant can only operate within their own organization.
// If an organization_id is present in the context, it overrides the requested domain.
func resolveDomainHTTP(r *http.Request, requestedDomain string) string {
	orgID, ok := delivery.GetContextString(r.Context(), delivery.OrganizationIDKey)
	if !ok || orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	if orgID != "" {
		return orgID
	}
	if requestedDomain == "" {
		return "global"
	}
	return requestedDomain
}

// actorContextHTTP derives the authenticated actor from the request context and
// propagates it into the usecase context so privilege guards can evaluate it.
func actorContextHTTP(r *http.Request) (context.Context, bool) {
	actorID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || actorID == "" {
		return nil, false
	}
	return authcontext.WithUserID(r.Context(), actorID), true
}

// HTTPAssignRole godoc
// @Summary      Assign role to user
// @Description  Assigns a role to a specified user (Casbin). Defaults to 'global' domain.
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID header string false "Organization ID"
// @Param        X-Organization-Slug header string false "Organization Slug"
// @Param        request body model.AssignRoleRequest true "Assign Role Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Role assigned successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/assign-role [post]
func (h *PermissionController) HTTPAssignRole(w http.ResponseWriter, r *http.Request) {
	var req model.AssignRoleRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	ctx, ok := actorContextHTTP(r)
	if !ok {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	err := h.useCase.AssignRoleToUser(ctx, req.UserID, req.Role, resolveDomainHTTP(r, req.Domain))
	if err != nil {
		response.WriteHTTPError(w, err, "failed to assign role")
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "role assigned successfully"})
}

// HTTPRevokeRole godoc
// @Summary      Revoke role from user
// @Description  Revokes a role from a specified user (Casbin). Defaults to 'global' domain.
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.AssignRoleRequest true "Revoke Role Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Role revoked successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/revoke-role [delete]
func (h *PermissionController) HTTPRevokeRole(w http.ResponseWriter, r *http.Request) {
	var req model.AssignRoleRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	err := h.useCase.RevokeRoleFromUser(r.Context(), req.UserID, req.Role, resolveDomainHTTP(r, req.Domain))
	if err != nil {
		if message, ok := usecase.IsNoopError(err); ok {
			respondPermissionNoopHTTP(w, message)
			return
		}
		response.WriteHTTPError(w, err, "failed to revoke role")
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "role revoked successfully"})
}

// HTTPGrantPermission godoc
// @Summary      Grant permission to role
// @Description  Grants a permission (path + method) to a role (Casbin). Defaults to 'global' domain.
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.GrantPermissionRequest true "Grant Permission Request"
// @Success      201  {object}  response.SwaggerGeneralResponseWrapper "Permission granted successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/grant [post]
func (h *PermissionController) HTTPGrantPermission(w http.ResponseWriter, r *http.Request) {
	var req model.GrantPermissionRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	ctx, ok := actorContextHTTP(r)
	if !ok {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	err := h.useCase.GrantPermissionToRole(ctx, req.Role, req.Path, req.Method, resolveDomainHTTP(r, req.Domain))
	if err != nil {
		if message, ok := usecase.IsNoopError(err); ok {
			respondPermissionNoopHTTP(w, message)
			return
		}
		response.WriteHTTPError(w, err, "failed to grant permission")
		return
	}
	response.WriteSuccess(w, http.StatusCreated, map[string]string{"message": "permission granted successfully"})
}

// HTTPGetAllPermissions godoc
// @Summary      Get all permissions
// @Description  Retrieves all Casbin policies.
// @Tags         permissions
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.SwaggerPermissionListResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions [get]
func (h *PermissionController) HTTPGetAllPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := h.useCase.GetAllPermissions(r.Context())
	if err != nil {
		response.WriteHTTPError(w, err, "failed to get all permissions")
		return
	}
	response.WriteSuccess(w, http.StatusOK, filterPoliciesByDomainHTTP(r, permissions, resolveDomainHTTP(r, "")))
}

// HTTPGetPermissionsForRole godoc
// @Summary      Get permissions for role
// @Description  Retrieves all permissions assigned to a specific role.
// @Tags         permissions
// @Security     BearerAuth
// @Produce      json
// @Param        role path string true "Role name"
// @Success      200  {object}  response.SwaggerPermissionListResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Role is required"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/{role} [get]
func (h *PermissionController) HTTPGetPermissionsForRole(w http.ResponseWriter, r *http.Request) {
	role := r.PathValue("role")
	if role == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "role is required")
		return
	}
	permissions, err := h.useCase.GetPermissionsForRole(r.Context(), role)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to get permissions for role")
		return
	}
	response.WriteSuccess(w, http.StatusOK, filterPoliciesByDomainHTTP(r, permissions, resolveDomainHTTP(r, "")))
}

// HTTPGetUsersForRole godoc
// @Summary      Get users for role
// @Description  Retrieves all user IDs assigned to a specific role. Defaults to 'global' domain.
// @Tags         permissions
// @Security     BearerAuth
// @Produce      json
// @Param        role path string true "Role name"
// @Param        domain query string false "Domain/Organization ID (defaults to 'global')"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "List of user IDs"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Role is required"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/roles/{role}/users [get]
func (h *PermissionController) HTTPGetUsersForRole(w http.ResponseWriter, r *http.Request) {
	role := r.PathValue("role")
	if role == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "role is required")
		return
	}
	domain := resolveDomainHTTP(r, r.URL.Query().Get("domain"))
	users, err := h.useCase.GetUsersForRole(r.Context(), role, domain)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to get users for role")
		return
	}
	response.WriteSuccess(w, http.StatusOK, users)
}

func filterPoliciesByDomainHTTP(r *http.Request, policies [][]string, domain string) [][]string {
	userRole, _ := delivery.GetContextString(r.Context(), delivery.RoleKey)
	isSuperAdmin := userRole == "role:superadmin"
	if (domain == "" || domain == "global") && isSuperAdmin {
		return policies
	}
	targetDomain := domain
	if targetDomain == "" || targetDomain == "global" {
		if orgID, ok := delivery.GetContextString(r.Context(), delivery.OrganizationIDKey); ok && orgID != "" {
			targetDomain = orgID
		} else if orgID := database.GetOrganizationID(r.Context()); orgID != "" {
			targetDomain = orgID
		}
	}
	filtered := make([][]string, 0, len(policies))
	for _, policy := range policies {
		if len(policy) > 1 && policy[1] == targetDomain {
			filtered = append(filtered, policy)
		}
	}
	return filtered
}

// HTTPUpdatePermission godoc
// @Summary      Update permission
// @Description  Updates an existing Casbin policy.
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.UpdatePermissionRequest true "Update Permission Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Permission updated successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions [put]
func (h *PermissionController) HTTPUpdatePermission(w http.ResponseWriter, r *http.Request) {
	var req model.UpdatePermissionRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	_, err := h.useCase.UpdatePermission(r.Context(), req.OldPermission, req.NewPermission)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to update permission")
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "permission updated successfully"})
}

// HTTPRevokePermission godoc
// @Summary      Revoke permission from role
// @Description  Revokes a permission (path + method) from a role (Casbin). Defaults to 'global' domain.
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.GrantPermissionRequest true "Revoke Permission Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Permission revoked successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/revoke [delete]
func (h *PermissionController) HTTPRevokePermission(w http.ResponseWriter, r *http.Request) {
	var req model.GrantPermissionRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	err := h.useCase.RevokePermissionFromRole(r.Context(), req.Role, req.Path, req.Method, resolveDomainHTTP(r, req.Domain))
	if err != nil {
		if message, ok := usecase.IsNoopError(err); ok {
			respondPermissionNoopHTTP(w, message)
			return
		}
		response.WriteHTTPError(w, err, "failed to revoke permission")
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "permission revoked successfully"})
}

// HTTPAddRoleInheritance godoc
// @Summary      Add role inheritance
// @Description  Creates a parent-child relationship between two roles. Defaults to 'global' domain.
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.RoleInheritanceRequest true "Role Inheritance Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Role inheritance added successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/inheritance [post]
func (h *PermissionController) HTTPAddRoleInheritance(w http.ResponseWriter, r *http.Request) {
	var req model.RoleInheritanceRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	ctx, ok := actorContextHTTP(r)
	if !ok {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	err := h.useCase.AddParentRole(ctx, req.ChildRole, req.ParentRole, resolveDomainHTTP(r, req.Domain))
	if err != nil {
		response.WriteHTTPError(w, err, "failed to add role inheritance")
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "role inheritance added successfully"})
}

// HTTPRemoveRoleInheritance godoc
// @Summary      Remove role inheritance
// @Description  Removes a parent-child relationship between two roles. Defaults to 'global' domain.
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.RoleInheritanceRequest true "Role Inheritance Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Role inheritance removed successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/inheritance [delete]
func (h *PermissionController) HTTPRemoveRoleInheritance(w http.ResponseWriter, r *http.Request) {
	var req model.RoleInheritanceRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	err := h.useCase.RemoveParentRole(r.Context(), req.ChildRole, req.ParentRole, resolveDomainHTTP(r, req.Domain))
	if err != nil {
		if message, ok := usecase.IsNoopError(err); ok {
			respondPermissionNoopHTTP(w, message)
			return
		}
		response.WriteHTTPError(w, err, "failed to remove role inheritance")
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "role inheritance removed successfully"})
}

// HTTPGetParentRoles godoc
// @Summary      Get parent roles
// @Description  Retrieves all parent roles for a given role. Defaults to 'global' domain.
// @Tags         permissions
// @Security     BearerAuth
// @Produce      json
// @Param        role path string true "Role name"
// @Param        domain query string false "Domain/Organization ID (defaults to 'global')"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "List of parent roles"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Role is required"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/{role}/parents [get]
func (h *PermissionController) HTTPGetParentRoles(w http.ResponseWriter, r *http.Request) {
	role := r.PathValue("role")
	if role == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "role is required")
		return
	}
	domain := resolveDomainHTTP(r, r.URL.Query().Get("domain"))
	parents, err := h.useCase.GetParentRoles(r.Context(), role, domain)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to get parent roles")
		return
	}
	response.WriteSuccess(w, http.StatusOK, parents)
}

// HTTPBatchCheck godoc
// @Summary      Batch check permissions
// @Description  Checks multiple permissions for the current user in a single request.
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.BatchPermissionCheckRequest true "Batch Check Request"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=model.BatchPermissionCheckResponse}
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/check-batch [post]
func (h *PermissionController) HTTPBatchCheck(w http.ResponseWriter, r *http.Request) {
	var req model.BatchPermissionCheckRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || userID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	results, err := h.useCase.BatchCheckPermission(r.Context(), userID, req.Items)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to batch check permissions")
		return
	}
	response.WriteSuccess(w, http.StatusOK, model.BatchPermissionCheckResponse{Results: results})
}

// HTTPGetResourceAggregation godoc
// @Summary      Get resource aggregation
// @Description  Retrieves permissions aggregated by resource with CRUD mapping
// @Tags         permissions
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=model.ResourceAggregationResponse} "Resource aggregation retrieved successfully"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/resources [get]
func (h *PermissionController) HTTPGetResourceAggregation(w http.ResponseWriter, r *http.Request) {
	result, err := h.useCase.GetResourceAggregation(r.Context())
	if err != nil {
		response.WriteHTTPError(w, err, "failed to get resource aggregation")
		return
	}
	response.WriteSuccess(w, http.StatusOK, result)
}

// HTTPGetInheritanceTree godoc
// @Summary      Get role inheritance tree
// @Description  Retrieves role hierarchy with inherited and effective permissions
// @Tags         permissions
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=model.InheritanceTreeResponse} "Inheritance tree retrieved successfully"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/inheritance-tree [get]
func (h *PermissionController) HTTPGetInheritanceTree(w http.ResponseWriter, r *http.Request) {
	result, err := h.useCase.GetInheritanceTree(r.Context())
	if err != nil {
		response.WriteHTTPError(w, err, "failed to get inheritance tree")
		return
	}
	response.WriteSuccess(w, http.StatusOK, result)
}

// HTTPGetRoleAccessRights godoc
// @Summary      Get access rights assignment status for a role
// @Description  Returns all access rights with is_assigned/is_partial flags for the given role
// @Tags         permissions
// @Security     BearerAuth
// @Produce      json
// @Param        role    path     string  true  "Role name"
// @Param        domain  query    string  false "Domain (default: global)"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=[]model.RoleAccessRightStatus}
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /permissions/roles/{role}/access-rights [get]
func (h *PermissionController) HTTPGetRoleAccessRights(w http.ResponseWriter, r *http.Request) {
	role := r.PathValue("role")
	domain := resolveDomainHTTP(r, r.URL.Query().Get("domain"))
	result, err := h.useCase.GetRoleAccessRights(r.Context(), role, domain)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to get role access rights")
		return
	}
	response.WriteSuccess(w, http.StatusOK, result)
}

// HTTPAssignAccessRight godoc
// @Summary      Bulk assign an access right to a role
// @Description  Grants all endpoints of the given access right to the specified role
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body     model.AssignAccessRightRequest  true  "Assign request"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /permissions/assign-access-right [post]
func (h *PermissionController) HTTPAssignAccessRight(w http.ResponseWriter, r *http.Request) {
	var req model.AssignAccessRightRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	req.Domain = resolveDomainHTTP(r, req.Domain)
	ctx, ok := actorContextHTTP(r)
	if !ok {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	if err := h.useCase.AssignAccessRight(ctx, req); err != nil {
		if message, ok := usecase.IsNoopError(err); ok {
			respondPermissionNoopHTTP(w, message)
			return
		}
		response.WriteHTTPError(w, err, "failed to assign access right")
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "access right assigned successfully"})
}

// HTTPRevokeAccessRight godoc
// @Summary      Bulk revoke an access right from a role
// @Description  Removes all endpoints of the given access right from the specified role
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body     model.AssignAccessRightRequest  true  "Revoke request"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /permissions/revoke-access-right [delete]
func (h *PermissionController) HTTPRevokeAccessRight(w http.ResponseWriter, r *http.Request) {
	var req model.AssignAccessRightRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	req.Domain = resolveDomainHTTP(r, req.Domain)
	if err := h.useCase.RevokeAccessRight(r.Context(), req); err != nil {
		if message, ok := usecase.IsNoopError(err); ok {
			respondPermissionNoopHTTP(w, message)
			return
		}
		response.WriteHTTPError(w, err, "failed to revoke access right")
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "access right revoked successfully"})
}
