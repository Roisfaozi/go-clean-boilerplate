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
