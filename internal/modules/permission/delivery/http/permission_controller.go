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

func (h *PermissionController) GetAllPermissions(c *gin.Context) {
	permissions, err := h.useCase.GetAllPermissions(c.Request.Context())
	if err != nil {
		response.HandleError(c, err, "failed to get all permissions")
		return
	}

	response.Success(c, permissions)
}

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
