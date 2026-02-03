package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type OrganizationController struct {
	OrgUseCase    usecase.OrganizationUseCase
	MemberUseCase usecase.OrganizationMemberUseCase
	Log           *logrus.Logger
	validate      *validator.Validate
}

func NewOrganizationController(
	orgUseCase usecase.OrganizationUseCase,
	memberUseCase usecase.OrganizationMemberUseCase,
	log *logrus.Logger,
	validate *validator.Validate,
) *OrganizationController {
	return &OrganizationController{
		OrgUseCase:    orgUseCase,
		MemberUseCase: memberUseCase,
		Log:           log,
		validate:      validate,
	}
}

// CreateOrganization creates a new organization
// @Summary      Create organization
// @Description  Creates a new organization with the current user as owner
// @Tags         organizations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      model.CreateOrganizationRequest  true  "Organization creation request"
// @Success      201      {object}  response.SwaggerSuccessResponseWrapper{data=model.OrganizationResponse}
// @Failure      400      {object}  response.SwaggerErrorResponseWrapper  "Validation error"
// @Failure      401      {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      409      {object}  response.SwaggerErrorResponseWrapper  "Slug already exists"
// @Failure      500      {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations [post]
func (ctrl *OrganizationController) CreateOrganization(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, nil, "user not authenticated")
		return
	}

	var request model.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := ctrl.validate.Struct(&request); err != nil {
		errorMsg := validation.FormatValidationErrors(err)
		response.ValidationError(c, err, errorMsg)
		return
	}

	result, err := ctrl.OrgUseCase.CreateOrganization(c.Request.Context(), userID.(string), &request)
	if err != nil {
		if err == exception.ErrConflict {
			response.ErrorResponse(c, 409, err, "organization slug already exists")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to create organization")
		response.InternalServerError(c, err, "failed to create organization")
		return
	}

	response.Created(c, result)
}

// GetOrganization retrieves an organization by ID
// @Summary      Get organization by ID
// @Description  Retrieves organization details by its ID
// @Tags         organizations
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=model.OrganizationResponse}
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper  "Organization not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/{id} [get]
func (ctrl *OrganizationController) GetOrganization(c *gin.Context) {
	orgID := c.Param("id")

	result, err := ctrl.OrgUseCase.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		if err == exception.ErrNotFound {
			response.NotFound(c, err, "organization not found")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to get organization")
		response.InternalServerError(c, err, "failed to get organization")
		return
	}

	response.Success(c, result)
}

// GetOrganizationBySlug retrieves an organization by slug
// @Summary      Get organization by slug
// @Description  Retrieves organization details by its slug
// @Tags         organizations
// @Security     BearerAuth
// @Produce      json
// @Param        slug  path      string  true  "Organization Slug"
// @Success      200   {object}  response.SwaggerSuccessResponseWrapper{data=model.OrganizationResponse}
// @Failure      401   {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      404   {object}  response.SwaggerErrorResponseWrapper  "Organization not found"
// @Failure      500   {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/slug/{slug} [get]
func (ctrl *OrganizationController) GetOrganizationBySlug(c *gin.Context) {
	slug := c.Param("slug")

	result, err := ctrl.OrgUseCase.GetOrganizationBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == exception.ErrNotFound {
			response.NotFound(c, err, "organization not found")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to get organization by slug")
		response.InternalServerError(c, err, "failed to get organization")
		return
	}

	response.Success(c, result)
}

// UpdateOrganization updates organization details
// @Summary      Update organization
// @Description  Updates organization details (name, settings)
// @Tags         organizations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                            true  "Organization ID"
// @Param        request  body      model.UpdateOrganizationRequest   true  "Update request"
// @Success      200      {object}  response.SwaggerSuccessResponseWrapper{data=model.OrganizationResponse}
// @Failure      400      {object}  response.SwaggerErrorResponseWrapper  "Validation error"
// @Failure      401      {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      404      {object}  response.SwaggerErrorResponseWrapper  "Organization not found"
// @Failure      500      {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/{id} [put]
func (ctrl *OrganizationController) UpdateOrganization(c *gin.Context) {
	orgID := c.Param("id")

	var request model.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := ctrl.validate.Struct(&request); err != nil {
		errorMsg := validation.FormatValidationErrors(err)
		response.ValidationError(c, err, errorMsg)
		return
	}

	result, err := ctrl.OrgUseCase.UpdateOrganization(c.Request.Context(), orgID, &request)
	if err != nil {
		if err == exception.ErrNotFound {
			response.NotFound(c, err, "organization not found")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to update organization")
		response.InternalServerError(c, err, "failed to update organization")
		return
	}

	response.Success(c, result)
}

// DeleteOrganization deletes an organization
// @Summary      Delete organization
// @Description  Deletes an organization (owner only)
// @Tags         organizations
// @Security     BearerAuth
// @Produce      json
// @Param        id  path      string  true  "Organization ID"
// @Success      200 {object}  response.SwaggerSuccessResponseWrapper
// @Failure      401 {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      403 {object}  response.SwaggerErrorResponseWrapper  "Forbidden (not owner)"
// @Failure      404 {object}  response.SwaggerErrorResponseWrapper  "Organization not found"
// @Failure      500 {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/{id} [delete]
func (ctrl *OrganizationController) DeleteOrganization(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, nil, "user not authenticated")
		return
	}

	orgID := c.Param("id")

	err := ctrl.OrgUseCase.DeleteOrganization(c.Request.Context(), orgID, userID.(string))
	if err != nil {
		if err == exception.ErrNotFound {
			response.NotFound(c, err, "organization not found")
			return
		}
		if err == exception.ErrForbidden {
			response.Forbidden(c, err, "only the owner can delete this organization")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to delete organization")
		response.InternalServerError(c, err, "failed to delete organization")
		return
	}

	response.Success(c, nil)
}

// GetMyOrganizations retrieves organizations for the current user
// @Summary      Get my organizations
// @Description  Retrieves all organizations the current user is a member of
// @Tags         organizations
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=model.UserOrganizationsResponse}
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/me [get]
func (ctrl *OrganizationController) GetMyOrganizations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, nil, "user not authenticated")
		return
	}

	result, err := ctrl.OrgUseCase.GetUserOrganizations(c.Request.Context(), userID.(string))
	if err != nil {
		ctrl.Log.WithError(err).Error("Failed to get user organizations")
		response.InternalServerError(c, err, "failed to get organizations")
		return
	}

	response.Success(c, result)
}
