package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type WebhookController struct {
	UseCase usecase.WebhookUseCase
}

func NewWebhookController(useCase usecase.WebhookUseCase) *WebhookController {
	return &WebhookController{
		UseCase: useCase,
	}
}

// Create handles webhook creation
// @Summary      Create a new outbound webhook
// @Description  Registers a new webhook URL to receive specified events.
// @Tags         webhooks
// @Accept       json
// @Produce      json
// @Param        request body model.CreateWebhookRequest true "Webhook Details"
// @Success      201  {object}  response.SwaggerWebhookResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      403  {object}  response.SwaggerErrorResponseWrapper "Forbidden"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /webhooks [post]
func (c *WebhookController) Create(ctx *gin.Context) {
	var req model.CreateWebhookRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, exception.ErrBadRequest, "Invalid request body")
		return
	}

	orgID, ok := middleware.GetOrganizationIDFromContext(ctx)
	if !ok {
		response.HandleError(ctx, exception.ErrInternalServer, "Internal server error")
		return
	}
	req.OrganizationID = orgID

	res, err := c.UseCase.Create(ctx.Request.Context(), req)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			msg := validation.FormatValidationErrors(valErrs)
			response.ValidationError(ctx, exception.ErrValidationError, msg)
			return
		}
		response.HandleError(ctx, err, "Failed to create webhook")
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, res)
}

// Update handles webhook updates
// @Summary      Update an existing webhook
// @Description  Updates the configuration of an existing webhook.
// @Tags         webhooks
// @Accept       json
// @Produce      json
// @Param        id path string true "Webhook ID"
// @Param        organization_id query string true "Organization ID"
// @Param        request body model.UpdateWebhookRequest true "Webhook Update Details"
// @Success      200  {object}  response.SwaggerWebhookResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body or missing organization_id"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "Webhook not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /webhooks/{id} [put]
func (c *WebhookController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	orgID, ok := middleware.GetOrganizationIDFromContext(ctx)
	if !ok {
		response.HandleError(ctx, exception.ErrInternalServer, "Internal server error")
		return
	}

	var req model.UpdateWebhookRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, exception.ErrBadRequest, "Invalid request body")
		return
	}

	res, err := c.UseCase.Update(ctx.Request.Context(), id, orgID, req)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			msg := validation.FormatValidationErrors(valErrs)
			response.ValidationError(ctx, exception.ErrValidationError, msg)
			return
		}
		response.HandleError(ctx, err, "Failed to update webhook")
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, res)
}

// Delete handles webhook deletion
// @Summary      Delete a webhook
// @Description  Deletes an outbound webhook configuration.
// @Tags         webhooks
// @Param        id path string true "Webhook ID"
// @Param        organization_id query string true "Organization ID"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Webhook deleted successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Missing organization_id"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "Webhook not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /webhooks/{id} [delete]
func (c *WebhookController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	orgID, ok := middleware.GetOrganizationIDFromContext(ctx)
	if !ok {
		response.HandleError(ctx, exception.ErrInternalServer, "Internal server error")
		return
	}

	if err := c.UseCase.Delete(ctx.Request.Context(), id, orgID); err != nil {
		response.HandleError(ctx, err, "Failed to delete webhook")
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, nil)
}

// FindByID retrieves a webhook by ID
// @Summary      Get webhook details
// @Description  Returns the details of a specific webhook.
// @Tags         webhooks
// @Produce      json
// @Param        id path string true "Webhook ID"
// @Param        organization_id query string true "Organization ID"
// @Success      200  {object}  response.SwaggerWebhookResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Missing organization_id"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "Webhook not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /webhooks/{id} [get]
func (c *WebhookController) FindByID(ctx *gin.Context) {
	id := ctx.Param("id")
	orgID, ok := middleware.GetOrganizationIDFromContext(ctx)
	if !ok {
		response.HandleError(ctx, exception.ErrInternalServer, "Internal server error")
		return
	}

	res, err := c.UseCase.FindByID(ctx.Request.Context(), id, orgID)
	if err != nil {
		response.HandleError(ctx, err, "Webhook not found")
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, res)
}

// FindByOrganization retrieves all webhooks for an organization
// @Summary      List organization webhooks
// @Description  Returns a list of all webhooks registered for the given organization.
// @Tags         webhooks
// @Produce      json
// @Param        organization_id query string true "Organization ID"
// @Success      200  {object}  response.SwaggerWebhookListResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Missing organization_id"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /webhooks [get]
func (c *WebhookController) FindByOrganization(ctx *gin.Context) {
	orgID, ok := middleware.GetOrganizationIDFromContext(ctx)
	if !ok {
		response.BadRequest(ctx, exception.ErrBadRequest, "organization context is required")
		return
	}

	res, err := c.UseCase.FindByOrganizationID(ctx.Request.Context(), orgID)
	if err != nil {
		response.HandleError(ctx, err, "Failed to retrieve webhooks")
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, res)
}

// GetLogs retrieves delivery logs for a specific webhook
// @Summary      Get webhook logs
// @Description  Returns a paginated list of delivery logs for a specific webhook.
// @Tags         webhooks
// @Produce      json
// @Param        id path string true "Webhook ID"
// @Param        organization_id query string true "Organization ID"
// @Param        limit query int false "Limit"
// @Param        offset query int false "Offset"
// @Success      200  {object}  response.SwaggerWebhookLogListResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Missing organization_id"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /webhooks/{id}/logs [get]
func (c *WebhookController) GetLogs(ctx *gin.Context) {
	id := ctx.Param("id")
	orgID, ok := middleware.GetOrganizationIDFromContext(ctx)
	if !ok {
		response.HandleError(ctx, exception.ErrInternalServer, "Internal server error")
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	res, err := c.UseCase.FindLogs(ctx.Request.Context(), id, orgID, limit, offset)
	if err != nil {
		response.HandleError(ctx, err, "Failed to retrieve webhook logs")
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, res)
}

func (c *WebhookController) HTTPCreate(w http.ResponseWriter, r *http.Request) {
	var req model.CreateWebhookRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "Invalid request body")
		return
	}
	orgID, ok := delivery.GetContextString(r.Context(), delivery.OrganizationIDKey)
	if !ok || orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	if orgID == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "organization context is required")
		return
	}
	req.OrganizationID = orgID
	res, err := c.UseCase.Create(r.Context(), req)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(valErrs))
			return
		}
		response.WriteHTTPError(w, err, "Failed to create webhook")
		return
	}
	response.WriteSuccess(w, http.StatusCreated, res)
}

func (c *WebhookController) HTTPUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	orgID, ok := delivery.GetContextString(r.Context(), delivery.OrganizationIDKey)
	if !ok || orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	if orgID == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "organization context is required")
		return
	}
	var req model.UpdateWebhookRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "Invalid request body")
		return
	}
	res, err := c.UseCase.Update(r.Context(), id, orgID, req)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(valErrs))
			return
		}
		response.WriteHTTPError(w, err, "Failed to update webhook")
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}

func (c *WebhookController) HTTPDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	orgID, ok := delivery.GetContextString(r.Context(), delivery.OrganizationIDKey)
	if !ok || orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	if orgID == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "organization context is required")
		return
	}
	if err := c.UseCase.Delete(r.Context(), id, orgID); err != nil {
		response.WriteHTTPError(w, err, "Failed to delete webhook")
		return
	}
	response.WriteSuccess(w, http.StatusOK, nil)
}

func (c *WebhookController) HTTPFindByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	orgID, ok := delivery.GetContextString(r.Context(), delivery.OrganizationIDKey)
	if !ok || orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	if orgID == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "organization context is required")
		return
	}
	res, err := c.UseCase.FindByID(r.Context(), id, orgID)
	if err != nil {
		response.WriteHTTPError(w, err, "Webhook not found")
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}

func (c *WebhookController) HTTPFindByOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, ok := delivery.GetContextString(r.Context(), delivery.OrganizationIDKey)
	if !ok || orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	if orgID == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "organization context is required")
		return
	}
	res, err := c.UseCase.FindByOrganizationID(r.Context(), orgID)
	if err != nil {
		response.WriteHTTPError(w, err, "Failed to retrieve webhooks")
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}

func (c *WebhookController) HTTPGetLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	orgID, ok := delivery.GetContextString(r.Context(), delivery.OrganizationIDKey)
	if !ok || orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	if orgID == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "organization context is required")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	res, err := c.UseCase.FindLogs(r.Context(), id, orgID, limit, offset)
	if err != nil {
		response.WriteHTTPError(w, err, "Failed to retrieve webhook logs")
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}
