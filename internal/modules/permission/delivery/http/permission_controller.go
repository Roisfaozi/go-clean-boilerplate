package http

import (
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

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

// AssignRole godoc
// @Summary      Assign role to user
// @Description  Assigns a role to a specified user (Casbin).
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.AssignRoleRequest true "Assign Role Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Role assigned successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/assign-role [post]
func (h *PermissionController) AssignRole(c *gin.Context) {
	var req model.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
		return
	}

	err := h.useCase.AssignRoleToUser(c.Request.Context(), req.UserID, req.Role)
	if err != nil {
		response.HandleError(c, err, "failed to assign role")
		return
	}

	response.Success(c, gin.H{"message": "role assigned successfully"})
}

// RevokeRole godoc
// @Summary      Revoke role from user
// @Description  Revokes a role from a specified user (Casbin).
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
// @Router       /permissions/revoke-role [post]
func (h *PermissionController) RevokeRole(c *gin.Context) {
	var req model.AssignRoleRequest // Same request structure as Assign
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
		return
	}

	err := h.useCase.RevokeRoleFromUser(c.Request.Context(), req.UserID, req.Role)
	if err != nil {
		response.HandleError(c, err, "failed to revoke role")
		return
	}

	response.Success(c, gin.H{"message": "role revoked successfully"})
}

// GrantPermission godoc
// @Summary      Grant permission to role
// @Description  Grants a permission (path + method) to a role (Casbin).
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
func (h *PermissionController) GrantPermission(c *gin.Context) {
	var req model.GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
		return
	}

	err := h.useCase.GrantPermissionToRole(c.Request.Context(), req.Role, req.Path, req.Method)
	if err != nil {
		response.HandleError(c, err, "failed to grant permission")
		return
	}

	response.Created(c, gin.H{"message": "permission granted successfully"})
}

// GetAllPermissions godoc
// @Summary      Get all permissions
// @Description  Retrieves all Casbin policies.
// @Tags         permissions
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.SwaggerPermissionListResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions [get]
func (h *PermissionController) GetAllPermissions(c *gin.Context) {
	permissions, err := h.useCase.GetAllPermissions(c.Request.Context())
	if err != nil {
		response.HandleError(c, err, "failed to get all permissions")
		return
	}

	response.Success(c, permissions)
}

// GetPermissionsForRole godoc
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
// @Router       /permissions/roles/{role} [get]
func (h *PermissionController) GetPermissionsForRole(c *gin.Context) {
	role := c.Param("role")
	if role == "" {
		response.BadRequest(c, nil, "role is required")
		return
	}

	permissions, err := h.useCase.GetPermissionsForRole(c.Request.Context(), role)
	if err != nil {
		response.HandleError(c, err, "failed to get permissions for role")
		return
	}

	response.Success(c, permissions)
}

// UpdatePermission godoc
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
func (h *PermissionController) UpdatePermission(c *gin.Context) {
	var req model.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
		return
	}

	_, err := h.useCase.UpdatePermission(c.Request.Context(), req.OldPermission, req.NewPermission)
	if err != nil {
		response.HandleError(c, err, "failed to update permission")
		return
	}

	response.Success(c, gin.H{"message": "permission updated successfully"})
}

// RevokePermission godoc
// @Summary      Revoke permission from role
// @Description  Revokes a permission (path + method) from a role (Casbin).
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
// @Router       /permissions/revoke [post]
func (h *PermissionController) RevokePermission(c *gin.Context) {
	var req model.GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
		return
	}

	err := h.useCase.RevokePermissionFromRole(c.Request.Context(), req.Role, req.Path, req.Method)
	if err != nil {
		response.HandleError(c, err, "failed to revoke permission")
		return
	}

	response.Success(c, gin.H{"message": "permission revoked successfully"})
}

// AddRoleInheritance godoc
// @Summary      Add role inheritance
// @Description  Creates a parent-child relationship between two roles.
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
func (h *PermissionController) AddRoleInheritance(c *gin.Context) {
	var req model.RoleInheritanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
		return
	}

	err := h.useCase.AddParentRole(c.Request.Context(), req.ChildRole, req.ParentRole)
	if err != nil {
		response.HandleError(c, err, "failed to add role inheritance")
		return
	}

	response.Success(c, gin.H{"message": "role inheritance added successfully"})
}

// RemoveRoleInheritance godoc
// @Summary      Remove role inheritance
// @Description  Removes a parent-child relationship between two roles.
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
func (h *PermissionController) RemoveRoleInheritance(c *gin.Context) {
	var req model.RoleInheritanceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
		return
	}

	err := h.useCase.RemoveParentRole(c.Request.Context(), req.ChildRole, req.ParentRole)
	if err != nil {
		response.HandleError(c, err, "failed to remove role inheritance")
		return
	}

	response.Success(c, gin.H{"message": "role inheritance removed successfully"})
}

// GetParentRoles godoc
// @Summary      Get parent roles
// @Description  Retrieves all parent roles for a given role.
// @Tags         permissions
// @Security     BearerAuth
// @Produce      json
// @Param        role path string true "Role name"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "List of parent roles"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Role is required"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /permissions/parents/{role} [get]
func (h *PermissionController) GetParentRoles(c *gin.Context) {
	role := c.Param("role")
	if role == "" {
		response.BadRequest(c, nil, "role is required")
		return
	}

	parents, err := h.useCase.GetParentRoles(c.Request.Context(), role)
	if err != nil {
		response.HandleError(c, err, "failed to get parent roles")
		return
	}

	response.Success(c, parents)
}

// BatchCheck godoc
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
func (h *PermissionController) BatchCheck(c *gin.Context) {
	var req model.BatchPermissionCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, errors.New("missing user id"), "user not authenticated")
		return
	}

	results, err := h.useCase.BatchCheckPermission(c.Request.Context(), userID.(string), req.Items)
	if err != nil {
		response.HandleError(c, err, "failed to batch check permissions")
		return
	}

	response.Success(c, model.BatchPermissionCheckResponse{Results: results})
}
