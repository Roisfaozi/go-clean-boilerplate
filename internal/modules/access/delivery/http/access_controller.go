package http

import (
	"errors"

	"github.com/Roisfaozi/casbin-db/internal/modules/access/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/usecase"
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

func (h *AccessHandler) CreateAccessRight(c *gin.Context) {
	var req model.CreateAccessRightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errors.New("invalid request body"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err)
		return
	}

	accessRight, err := h.useCase.CreateAccessRight(c.Request.Context(), req)
	if err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			response.ValidationError(c, err)
			return
		}
		h.log.WithError(err).Error("Failed to create access right")
		response.InternalServerError(c, errors.New("could not create access right"))
		return
	}

	response.Created(c, accessRight)
}

func (h *AccessHandler) GetAllAccessRights(c *gin.Context) {
	accessRights, err := h.useCase.GetAllAccessRights(c.Request.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed to get all access rights")
		response.InternalServerError(c, errors.New("could not retrieve access rights"))
		return
	}

	response.Success(c, accessRights)
}

func (h *AccessHandler) CreateEndpoint(c *gin.Context) {
	var req model.CreateEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errors.New("invalid request body"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err)
		return
	}

	endpoint, err := h.useCase.CreateEndpoint(c.Request.Context(), req)
	if err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			response.ValidationError(c, err)
			return
		}
		h.log.WithError(err).Error("Failed to create endpoint")
		response.InternalServerError(c, errors.New("could not create endpoint"))
		return
	}

	response.Created(c, endpoint)
}

func (h *AccessHandler) LinkEndpointToAccessRight(c *gin.Context) {
	var req model.LinkEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errors.New("invalid request body"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err)
		return
	}

	err := h.useCase.LinkEndpointToAccessRight(c.Request.Context(), req)
	if err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			response.ValidationError(c, err)
			return
		}
		h.log.WithError(err).Error("Failed to link endpoint to access right")
		response.InternalServerError(c, errors.New("could not link endpoint"))
		return
	}

	response.Success(c, gin.H{"message": "Endpoint linked successfully"})
}
