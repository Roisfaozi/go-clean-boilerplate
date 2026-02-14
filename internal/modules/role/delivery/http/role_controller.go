package http

import (
	"errors"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type RoleController struct {
	RoleUseCase usecase.RoleUseCase
	Log         *logrus.Logger
	validate    *validator.Validate
}

func NewRoleController(roleUseCase usecase.RoleUseCase, log *logrus.Logger, validate *validator.Validate) *RoleController {
	return &RoleController{
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
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      409  {object}  response.SwaggerErrorResponseWrapper "Role already exists"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /roles [post]
func (h *RoleController) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.WithError(err).Error("failed to bind request body for create role")
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, exception.ErrValidationError, msg)
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
func (h *RoleController) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	roles, err := h.RoleUseCase.GetAll(ctx)
	if err != nil {
		h.handleError(c, err, "failed to get all roles")
		return
	}

	response.Success(c, roles)
}

// Update updates a role
// @Summary      Update role
// @Description  Update a role by ID.
// @Tags         roles
// @Security     BearerAuth
// @Param        id   path      string  true  "Role ID"
// @Param        request body model.UpdateRoleRequest true "Role Update Details"
// @Produce      json
// @Success      200  {object}  response.SwaggerRoleResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "Role not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /roles/{id} [put]
func (h *RoleController) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	var req model.UpdateRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.WithError(err).Error("failed to bind request body for update role")
		response.BadRequest(c, err, "invalid request body")
		return
	}

	req.Sanitize()

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	role, err := h.RoleUseCase.Update(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "failed to update role")
		return
	}

	response.Success(c, role)
}

// Delete removes a role
// @Summary      Delete role
// @Description  Deletes a role by ID. Only superadmin should have access.
// @Tags         roles
// @Security     BearerAuth
// @Param        id   path      string  true  "Role ID"
// @Produce      json
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Role deleted successfully"
// @Failure      403  {object}  response.SwaggerErrorResponseWrapper "Forbidden (cannot delete superadmin)"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "Role not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /roles/{id} [delete]
func (h *RoleController) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.RoleUseCase.Delete(ctx, id); err != nil {
		h.handleError(c, err, "failed to delete role")
		return
	}

	response.Success(c, gin.H{"message": "Role deleted successfully"})
}

// GetRolesDynamic retrieves roles based on dynamic filters and sorting via POST request body
// @Summary      Get roles with dynamic filters
// @Description  Retrieves a list of roles based on dynamic filter and sort criteria provided in the request body.
// @Tags         roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        filter body querybuilder.DynamicFilter true "Dynamic filter and sort criteria"
// @Success      200  {object}  response.SwaggerRoleListResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body or filter criteria"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      403  {object}  response.SwaggerErrorResponseWrapper "Forbidden"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /roles/search [post]
func (h *RoleController) GetRolesDynamic(c *gin.Context) {
	ctx := c.Request.Context()
	var filter querybuilder.DynamicFilter

	if err := c.ShouldBindJSON(&filter); err != nil {
		h.Log.WithError(err).Error("failed to bind dynamic filter request body for roles")
		response.BadRequest(c, err, "invalid request body for dynamic filter")
		return
	}

	if err := h.validate.Struct(filter); err != nil {
		msg := validation.FormatValidationErrors(err)
		h.Log.WithError(err).Error(msg)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	roles, err := h.RoleUseCase.GetAllRolesDynamic(ctx, &filter)
	if err != nil {
		h.Log.WithError(err).Error("failed to get roles dynamically")
		h.handleError(c, err, "failed to retrieve roles")
		return
	}

	response.Success(c, roles)
}

func (h *RoleController) handleError(c *gin.Context, err error, message string) {
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
