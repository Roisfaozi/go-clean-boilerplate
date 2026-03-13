package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
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
		response.ErrorResponse(ctx, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	res, err := c.UseCase.Create(ctx.Request.Context(), req)
	if err != nil {
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
	orgID := ctx.Query("organization_id")
	if orgID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, errors.New("missing organization_id"), "organization_id is required")
		return
	}

	var req model.UpdateWebhookRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	res, err := c.UseCase.Update(ctx.Request.Context(), id, orgID, req)
	if err != nil {
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
	orgID := ctx.Query("organization_id")
	if orgID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, errors.New("missing organization_id"), "organization_id is required")
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
	orgID := ctx.Query("organization_id")
	if orgID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, errors.New("missing organization_id"), "organization_id is required")
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
	orgID := ctx.Query("organization_id")
	if orgID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, errors.New("missing organization_id"), "organization_id is required")
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
	orgID := ctx.Query("organization_id")
	if orgID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, errors.New("missing organization_id"), "organization_id is required")
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
