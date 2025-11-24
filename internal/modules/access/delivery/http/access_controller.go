package http

import (
	"errors"

	"github.com/Roisfaozi/casbin-db/internal/modules/access/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/Roisfaozi/casbin-db/internal/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type AccessHandler struct {
	useCase  usecase.IAccessUseCase
	validate *validator.Validate
	log      *logrus.Logger
}

func NewAccessHandler(useCase usecase.IAccessUseCase, validate *validator.Validate, log *logrus.Logger) *AccessHandler {
	return &AccessHandler{
		useCase:  useCase,
		validate: validate,
		log:      log,
	}
}

// @Summary      Create access right
// @Description  Creates a new access right (resource group).
// @Tags         access-rights
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.CreateAccessRightRequest true "Create Access Right Request"
// @Success      201  {object}  response.SwaggerAccessRightResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /access-rights [post]
func (h *AccessHandler) CreateAccessRight(c *gin.Context) {
	var req model.CreateAccessRightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response.ValidationError(c, exception.ErrBadRequest, validationErrors.Error())
		return
	}

	accessRight, err := h.useCase.CreateAccessRight(c.Request.Context(), req)
	if err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			response.ValidationError(c, exception.ErrBadRequest, err.Error())
			return
		}
		h.log.WithError(err).Error("Failed to create access right")
		response.InternalServerError(c, errors.New("could not create access right"), "failed to create access right")
		return
	}

	response.Created(c, accessRight)
}

// @Summary      List all access rights
// @Description  Retrieves a list of all available access rights.
// @Tags         access-rights
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.SwaggerAccessRightListResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /access-rights [get]
func (h *AccessHandler) GetAllAccessRights(c *gin.Context) {
	accessRights, err := h.useCase.GetAllAccessRights(c.Request.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed to get all access rights")
		response.InternalServerError(c, errors.New("could not retrieve access rights"), "failed to get all access rights")
		return
	}

	response.Success(c, accessRights)
}

// @Summary      Create endpoint
// @Description  Registers a new API endpoint in the system.
// @Tags         endpoints
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.CreateEndpointRequest true "Create Endpoint Request"
// @Success      201  {object}  response.SwaggerEndpointResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /endpoints [post]
func (h *AccessHandler) CreateEndpoint(c *gin.Context) {
	var req model.CreateEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response.ValidationError(c, exception.ErrBadRequest, validationErrors.Error())
		return
	}

	endpoint, err := h.useCase.CreateEndpoint(c.Request.Context(), req)
	if err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			response.ValidationError(c, exception.ErrBadRequest, err.Error())
			return
		}
		h.log.WithError(err).Error("Failed to create endpoint")
		response.InternalServerError(c, errors.New("could not create endpoint"), "failed to create endpoint")
		return
	}

	response.Created(c, endpoint)
}

// @Summary      Link endpoint to access right
// @Description  Associates an endpoint with a specific access right.
// @Tags         access-rights
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.LinkEndpointRequest true "Link Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Endpoint linked successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /access-rights/link [post]
func (h *AccessHandler) LinkEndpointToAccessRight(c *gin.Context) {
	var req model.LinkEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		response.ValidationError(c, exception.ErrBadRequest, validationErrors.Error())
		return
	}

	err := h.useCase.LinkEndpointToAccessRight(c.Request.Context(), req)
	if err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			response.ValidationError(c, exception.ErrBadRequest, err.Error())
			return
		}
		h.log.WithError(err).Error("Failed to link endpoint to access right")
		response.InternalServerError(c, errors.New("could not link endpoint"), "failed to link endpoint")
		return
	}

	response.Success(c, gin.H{"message": "Endpoint linked successfully"})
}