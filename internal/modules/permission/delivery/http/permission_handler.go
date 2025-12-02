package http

import (
	"errors"

	"github.com/Roisfaozi/casbin-db/internal/modules/permission/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/permission/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/Roisfaozi/casbin-db/internal/utils/response"
	"github.com/Roisfaozi/casbin-db/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type PermissionHandler struct {
	useCase  usecase.IPermissionUseCase
	validate *validator.Validate
	log      *logrus.Logger
}

func NewPermissionHandler(useCase usecase.IPermissionUseCase, validate *validator.Validate, log *logrus.Logger) *PermissionHandler {
	return &PermissionHandler{
		useCase:  useCase,
		validate: validate,
		log:      log,
	}
}

// AssignRole assigns a role to a user
// @Summary      Assign role to user
// @Description  Assigns a specific role to a user.
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.AssignRoleRequest true "Assign Role Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Role assigned successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/assign-role [post]
func (h *PermissionHandler) AssignRole(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	if err := h.useCase.AssignRoleToUser(ctx, req.UserID, req.Role); err != nil {
		h.log.WithError(err).Error("Failed to assign role")
		response.InternalServerError(c, errors.New("could not assign role"), "failed to assign role")
		return
	}

	response.Success(c, gin.H{"message": "Role assigned successfully"})
}

// GrantPermission grants a permission to a role
// @Summary      Grant permission
// @Description  Grants a specific permission (path & method) to a role.
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.GrantPermissionRequest true "Grant Permission Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Permission granted successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/grant [post]
func (h *PermissionHandler) GrantPermission(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	if err := h.useCase.GrantPermissionToRole(ctx, req.Role, req.Path, req.Method); err != nil {
		h.log.WithError(err).Error("Failed to grant permission")
		response.InternalServerError(c, errors.New("could not grant permission"), "failed to grant permission")
		return
	}

	response.Success(c, gin.H{"message": "Permission granted successfully"})
}

// GetAllPermissions retrieves all permissions
// @Summary      Get all permissions
// @Description  Retrieves all policies from the Casbin enforcer.
// @Tags         permissions
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.SwaggerPermissionListResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions [get]
func (h *PermissionHandler) GetAllPermissions(c *gin.Context) {
	permissions, err := h.useCase.GetAllPermissions()
	if err != nil {
		h.log.WithError(err).Error("Failed to get all permissions")
		response.InternalServerError(c, errors.New("could not retrieve permissions"), "failed to get all permissions")
		return
	}
	response.Success(c, permissions)
}

// GetPermissionsForRole retrieves permissions for a specific role
// @Summary      Get permissions for role
// @Description  Retrieves all permissions associated with a specific role.
// @Tags         permissions
// @Security     BearerAuth
// @Produce      json
// @Param        role path string true "Role Name"
// @Success      200  {object}  response.SwaggerPermissionListResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Role parameter required"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/{role} [get]
func (h *PermissionHandler) GetPermissionsForRole(c *gin.Context) {
	role := c.Param("role")
	if role == "" {
		response.BadRequest(c, exception.ErrBadRequest, "role parameter is required")
		return
	}

	permissions, err := h.useCase.GetPermissionsForRole(role)
	if err != nil {
		h.log.WithError(err).Error("Failed to get permissions for role")
		response.InternalServerError(c, errors.New("could not retrieve permissions for role"), "failed to get permissions for role")
		return
	}
	response.Success(c, permissions)
}

// UpdatePermission updates an existing permission
// @Summary      Update permission
// @Description  Updates an existing policy rule.
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.UpdatePermissionRequest true "Update Permission Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Permission updated successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "Policy to update not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions [put]
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	var req model.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	_, err := h.useCase.UpdatePermission(req.OldPermission, req.NewPermission)
	if err != nil {
		h.log.WithError(err).Error("Failed to update permission")
		if err.Error() == "policy to update not found" {
			response.NotFound(c, err, "policy to update not found")
			return
		}
		response.InternalServerError(c, errors.New("could not update permission"), "failed to update permission")
		return
	}

	response.Success(c, gin.H{"message": "Permission updated successfully"})
}

// RevokePermission revokes a permission from a role
// @Summary      Revoke permission
// @Description  Revokes a specific permission (path & method) from a role.
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.GrantPermissionRequest true "Revoke Permission Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Permission revoked successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/revoke [delete]
func (h *PermissionHandler) RevokePermission(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.GrantPermissionRequest // Reuse GrantPermissionRequest as it has Role, Path, Method
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	if err := h.useCase.RevokePermissionFromRole(ctx, req.Role, req.Path, req.Method); err != nil {
		h.log.WithError(err).Error("Failed to revoke permission")
		response.InternalServerError(c, errors.New("could not revoke permission"), "failed to revoke permission")
		return
	}

	response.Success(c, gin.H{"message": "Permission revoked successfully"})
}