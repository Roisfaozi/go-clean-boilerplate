package http

import (
	"errors"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
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

// HTTPCreate creates a new role
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
func (h *RoleController) HTTPCreate(w http.ResponseWriter, r *http.Request) {
	var req model.CreateRoleRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		h.Log.WithError(err).Error("failed to bind request body for create role")
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	req.Sanitize()
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	role, err := h.RoleUseCase.Create(r.Context(), &req)
	if err != nil {
		h.handleHTTPError(w, err, "failed to create role")
		return
	}
	response.WriteSuccess(w, http.StatusCreated, role)
}

// HTTPGetAll lists all roles
// @Summary      List all roles
// @Description  Get a list of all available roles.
// @Tags         roles
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.SwaggerRoleListResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /roles [get]
func (h *RoleController) HTTPGetAll(w http.ResponseWriter, r *http.Request) {
	roles, err := h.RoleUseCase.GetAll(r.Context())
	if err != nil {
		h.handleHTTPError(w, err, "failed to get all roles")
		return
	}
	response.WriteSuccess(w, http.StatusOK, roles)
}

// HTTPUpdate updates a role
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
func (h *RoleController) HTTPUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req model.UpdateRoleRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		h.Log.WithError(err).Error("failed to bind request body for update role")
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	req.Sanitize()
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	role, err := h.RoleUseCase.Update(r.Context(), id, &req)
	if err != nil {
		h.handleHTTPError(w, err, "failed to update role")
		return
	}
	response.WriteSuccess(w, http.StatusOK, role)
}

// HTTPDelete removes a role
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
func (h *RoleController) HTTPDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.RoleUseCase.Delete(r.Context(), id); err != nil {
		h.handleHTTPError(w, err, "failed to delete role")
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Role deleted successfully"})
}

// HTTPGetRolesDynamic retrieves roles based on dynamic filters and sorting via POST request body
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
func (h *RoleController) HTTPGetRolesDynamic(w http.ResponseWriter, r *http.Request) {
	var filter querybuilder.DynamicFilter
	if err := response.DecodeJSON(r, &filter, 1024*1024); err != nil {
		h.Log.WithError(err).Error("failed to bind dynamic filter request body for roles")
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body for dynamic filter")
		return
	}
	if err := h.validate.Struct(filter); err != nil {
		msg := validation.FormatValidationErrors(err)
		h.Log.WithError(err).Error(msg)
		response.WriteHTTPError(w, exception.ErrValidationError, msg)
		return
	}
	roles, err := h.RoleUseCase.GetAllRolesDynamic(r.Context(), &filter)
	if err != nil {
		h.Log.WithError(err).Error("failed to get roles dynamically")
		h.handleHTTPError(w, err, "failed to retrieve roles")
		return
	}
	response.WriteSuccess(w, http.StatusOK, roles)
}

// HTTPCreateOrganizationRole creates a custom role for an organization
// @Summary      Create organization custom role
// @Description  Creates a new organization-scoped role.
// @Tags         organization-roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                  true  "Organization ID"
// @Param        request  body      model.CreateRoleRequest true  "Role creation request"
// @Success      201      {object}  response.SwaggerRoleResponseWrapper
// @Failure      400      {object}  response.SwaggerErrorResponseWrapper
// @Failure      409      {object}  response.SwaggerErrorResponseWrapper
// @Failure      500      {object}  response.SwaggerErrorResponseWrapper
// @Router       /organizations/{id}/roles [post]
func (h *RoleController) HTTPCreateOrganizationRole(w http.ResponseWriter, r *http.Request) {
	orgID, _ := delivery.GetContextString(r.Context(), delivery.OrganizationIDKey)
	if orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	if orgID == "" {
		orgID = r.PathValue("id")
	}
	if orgID == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "organization context required")
		return
	}
	var req model.CreateRoleRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	req.Sanitize()
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	role, err := h.RoleUseCase.CreateForOrganization(r.Context(), orgID, &req)
	if err != nil {
		h.handleHTTPError(w, err, "failed to create organization role")
		return
	}
	response.WriteSuccess(w, http.StatusCreated, role)
}

// HTTPGetOrganizationRoles lists all custom roles for an organization
// @Summary      List organization custom roles
// @Description  Get a list of custom roles belonging to the organization.
// @Tags         organization-roles
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  response.SwaggerRoleListResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /organizations/{id}/roles [get]
func (h *RoleController) HTTPGetOrganizationRoles(w http.ResponseWriter, r *http.Request) {
	orgID, _ := delivery.GetContextString(r.Context(), delivery.OrganizationIDKey)
	if orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	if orgID == "" {
		orgID = r.PathValue("id")
	}
	if orgID == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "organization context required")
		return
	}
	roles, err := h.RoleUseCase.GetOrganizationRoles(r.Context(), orgID)
	if err != nil {
		h.handleHTTPError(w, err, "failed to get organization roles")
		return
	}
	response.WriteSuccess(w, http.StatusOK, roles)
}

// HTTPUpdateOrganizationRole updates a custom role of an organization
// @Summary      Update organization custom role
// @Description  Updates an existing custom role for an organization by role ID.
// @Tags         organization-roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                  true  "Organization ID"
// @Param        roleId   path      string                  true  "Role ID"
// @Param        request  body      model.UpdateRoleRequest true  "Role update request"
// @Success      200      {object}  response.SwaggerRoleResponseWrapper
// @Failure      400      {object}  response.SwaggerErrorResponseWrapper
// @Failure      404      {object}  response.SwaggerErrorResponseWrapper
// @Failure      500      {object}  response.SwaggerErrorResponseWrapper
// @Router       /organizations/{id}/roles/{roleId} [put]
func (h *RoleController) HTTPUpdateOrganizationRole(w http.ResponseWriter, r *http.Request) {
	orgID, _ := delivery.GetContextString(r.Context(), delivery.OrganizationIDKey)
	if orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	if orgID == "" {
		orgID = r.PathValue("id")
	}
	roleID := r.PathValue("roleId")
	if orgID == "" || roleID == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "organization and role ID required")
		return
	}
	var req model.UpdateRoleRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	req.Sanitize()
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	role, err := h.RoleUseCase.UpdateForOrganization(r.Context(), orgID, roleID, &req)
	if err != nil {
		h.handleHTTPError(w, err, "failed to update organization role")
		return
	}
	response.WriteSuccess(w, http.StatusOK, role)
}

// HTTPDeleteOrganizationRole deletes a custom role belonging to an organization
// @Summary      Delete organization custom role
// @Description  Deletes a custom role belonging to an organization.
// @Tags         organization-roles
// @Security     BearerAuth
// @Produce      json
// @Param        id       path      string  true  "Organization ID"
// @Param        roleId   path      string  true  "Role ID"
// @Success      200      {object}  response.SwaggerGeneralResponseWrapper "Role deleted successfully"
// @Failure      400      {object}  response.SwaggerErrorResponseWrapper
// @Failure      404      {object}  response.SwaggerErrorResponseWrapper
// @Failure      500      {object}  response.SwaggerErrorResponseWrapper
// @Router       /organizations/{id}/roles/{roleId} [delete]
func (h *RoleController) HTTPDeleteOrganizationRole(w http.ResponseWriter, r *http.Request) {
	orgID, _ := delivery.GetContextString(r.Context(), delivery.OrganizationIDKey)
	if orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	if orgID == "" {
		orgID = r.PathValue("id")
	}
	roleID := r.PathValue("roleId")
	if orgID == "" || roleID == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "organization and role ID required")
		return
	}
	if err := h.RoleUseCase.DeleteForOrganization(r.Context(), orgID, roleID); err != nil {
		h.handleHTTPError(w, err, "failed to delete organization role")
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Role deleted successfully"})
}

func (h *RoleController) handleHTTPError(w http.ResponseWriter, err error, message string) {
	h.Log.WithError(err).Error(message)
	switch {
	case errors.Is(err, exception.ErrBadRequest):
		response.WriteError(w, http.StatusBadRequest, err, message)
	case errors.Is(err, exception.ErrUnauthorized):
		response.WriteError(w, http.StatusUnauthorized, err, message)
	case errors.Is(err, exception.ErrForbidden):
		response.WriteError(w, http.StatusForbidden, err, message)
	case errors.Is(err, exception.ErrNotFound):
		response.WriteError(w, http.StatusNotFound, err, message)
	case errors.Is(err, exception.ErrConflict):
		response.WriteError(w, http.StatusConflict, err, message)
	default:
		response.WriteError(w, http.StatusInternalServerError, err, message)
	}
}
