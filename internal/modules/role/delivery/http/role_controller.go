package http

import (
	"errors"
	"net/http"

	"github.com/Roisfaozi/casbin-db/internal/modules/role/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/Roisfaozi/casbin-db/internal/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type RoleHandler struct {
	RoleUseCase usecase.RoleUseCase
	Log         *logrus.Logger
	validate    *validator.Validate
}

// NewRoleHandler creates a new RoleHandler instance.
//
// It takes the following parameters:
// - roleUseCase: the RoleUseCase implementation.
// - log: the logrus.Logger implementation.
// - validate: the validator.Validate implementation.
//
// It returns a pointer to the newly created RoleHandler.
func NewRoleHandler(roleUseCase usecase.RoleUseCase, log *logrus.Logger, validate *validator.Validate) *RoleHandler {
	return &RoleHandler{
		RoleUseCase: roleUseCase,
		Log:         log,
		validate:    validate,
	}
}

// Create creates a new role
// @Summary      Create a new role
// @Description  Create a new user role.
// @Tags         roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.CreateRoleRequest true "Role Creation Details"
// @Success      201  {object}  response.SwaggerRoleResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      409  {object}  response.SwaggerErrorResponseWrapper "Role already exists"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.WithError(err).Error("failed to bind request body for create role")
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response.ValidationError(c, exception.ErrBadRequest, validationErrors.Error())
		return
	}

	role, err := h.RoleUseCase.Create(ctx, &req)
	if err != nil {
		h.handleError(c, err, "failed to create role")
		return
	}

	response.Created(c, role)
}

// GetAll lists all roles
// @Summary      List all roles
// @Description  Get a list of all available roles.
// @Tags         roles
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.SwaggerRoleListResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /roles [get]
func (h *RoleHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	roles, err := h.RoleUseCase.GetAll(ctx)
	if err != nil {
		h.handleError(c, err, "failed to get all roles")
		return
	}

	response.Success(c, roles)
}

func (h *RoleHandler) handleError(c *gin.Context, err error, message string) {
	h.Log.WithError(err).Error(message)

	switch {
	case errors.Is(err, exception.ErrBadRequest):
		response.BadRequest(c, err, message)
	case errors.Is(err, exception.ErrUnauthorized):
		response.Unauthorized(c, err, message)
	case errors.Is(err, exception.ErrForbidden):
		response.Forbidden(c, err, message)
	case errors.Is(err, exception.ErrNotFound):
		response.NotFound(c, err, message)
	case errors.Is(err, exception.ErrConflict):
		response.ErrorResponse(c, http.StatusConflict, err, message)
	default:
		response.InternalServerError(c, exception.ErrInternalServer, "something went wrong")
	}
}
