package http

import (
	"errors"

	"github.com/Roisfaozi/casbin-db/internal/modules/permission/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/permission/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/response"
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

func (h *PermissionHandler) AssignRole(c *gin.Context) {
	var req model.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errors.New("invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.useCase.AssignRoleToUser(req.UserID, req.Role); err != nil {
		h.log.WithError(err).Error("Failed to assign role")
		response.InternalServerError(c, errors.New("could not assign role"))
		return
	}

	response.Success(c, gin.H{"message": "Role assigned successfully"})
}

func (h *PermissionHandler) GrantPermission(c *gin.Context) {
	var req model.GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errors.New("invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.useCase.GrantPermissionToRole(req.Role, req.Path, req.Method); err != nil {
		h.log.WithError(err).Error("Failed to grant permission")
		response.InternalServerError(c, errors.New("could not grant permission"))
		return
	}

	response.Success(c, gin.H{"message": "Permission granted successfully"})
}

// GetAllPermissions handles the request to get all permissions.
func (h *PermissionHandler) GetAllPermissions(c *gin.Context) {
	permissions, err := h.useCase.GetAllPermissions()
	if err != nil {
		h.log.WithError(err).Error("Failed to get all permissions")
		response.InternalServerError(c, errors.New("could not retrieve permissions"))
		return
	}
	response.Success(c, permissions)
}

// GetPermissionsForRole handles the request to get permissions for a specific role.
func (h *PermissionHandler) GetPermissionsForRole(c *gin.Context) {
	role := c.Param("role")
	if role == "" {
		response.BadRequest(c, errors.New("role parameter is required"))
		return
	}

	permissions, err := h.useCase.GetPermissionsForRole(role)
	if err != nil {
		h.log.WithError(err).Error("Failed to get permissions for role")
		response.InternalServerError(c, errors.New("could not retrieve permissions for role"))
		return
	}
	response.Success(c, permissions)
}

// UpdatePermission handles the request to update a permission.
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	var req model.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errors.New("invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err)
		return
	}

	_, err := h.useCase.UpdatePermission(req.OldPermission, req.NewPermission)
	if err != nil {
		h.log.WithError(err).Error("Failed to update permission")
		if err.Error() == "policy to update not found" {
			response.NotFound(c, err)
			return
		}
		response.InternalServerError(c, errors.New("could not update permission"))
		return
	}

	response.Success(c, gin.H{"message": "Permission updated successfully"})
}
