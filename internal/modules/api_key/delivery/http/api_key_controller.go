package http

import (
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type ApiKeyController struct {
	useCase   usecase.ApiKeyUseCase
	log       *logrus.Logger
	validator *validator.Validate
}

func NewApiKeyController(useCase usecase.ApiKeyUseCase, log *logrus.Logger, validator *validator.Validate) *ApiKeyController {
	return &ApiKeyController{
		useCase:   useCase,
		log:       log,
		validator: validator,
	}
}

func (h *ApiKeyController) Create(c *gin.Context) {
	var req model.CreateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.ValidationError(c, err, "validation failed")
		return
	}

	userID, _ := middleware.GetUserIDFromContext(c)
	orgID := c.GetString("organization_id")
	if orgID == "" {
		response.BadRequest(c, errors.New("organization context required"), "missing organization_id")
		return
	}

	res, err := h.useCase.Create(c.Request.Context(), userID, orgID, &req)
	if err != nil {
		response.InternalServerError(c, err, "failed to create api key")
		return
	}

	response.Created(c, res)
}

func (h *ApiKeyController) List(c *gin.Context) {
	orgID := c.GetString("organization_id")
	if orgID == "" {
		response.BadRequest(c, errors.New("organization context required"), "missing organization_id")
		return
	}

	res, err := h.useCase.List(c.Request.Context(), orgID)
	if err != nil {
		response.InternalServerError(c, err, "failed to list api keys")
		return
	}

	response.Success(c, res)
}

func (h *ApiKeyController) Revoke(c *gin.Context) {
	id := c.Param("id")
	orgID := c.GetString("organization_id")
	if orgID == "" {
		response.BadRequest(c, errors.New("organization context required"), "missing organization_id")
		return
	}

	err := h.useCase.Revoke(c.Request.Context(), orgID, id)
	if err != nil {
		response.HandleError(c, err, "failed to revoke api key")
		return
	}

	response.Success(c, nil)
}
