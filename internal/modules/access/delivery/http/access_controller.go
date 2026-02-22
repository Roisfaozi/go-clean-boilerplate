package http

import (
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type AccessController struct {
	useCase  usecase.IAccessUseCase
	validate *validator.Validate
	log      *logrus.Logger
}

func NewAccessController(useCase usecase.IAccessUseCase, validate *validator.Validate, log *logrus.Logger) *AccessController {
	return &AccessController{
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
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /access-rights [post]
func (h *AccessController) CreateAccessRight(c *gin.Context) {
	var req model.CreateAccessRightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	accessRight, err := h.useCase.CreateAccessRight(c.Request.Context(), req)
	if err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			msg := validation.FormatValidationErrors(err)
			response.ValidationError(c, exception.ErrValidationError, msg)
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
func (h *AccessController) GetAllAccessRights(c *gin.Context) {
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
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /endpoints [post]
func (h *AccessController) CreateEndpoint(c *gin.Context) {
	var req model.CreateEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	endpoint, err := h.useCase.CreateEndpoint(c.Request.Context(), req)
	if err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			msg := validation.FormatValidationErrors(err)
			response.ValidationError(c, exception.ErrValidationError, msg)
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
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /access-rights/link [post]
func (h *AccessController) LinkEndpointToAccessRight(c *gin.Context) {
	var req model.LinkEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	err := h.useCase.LinkEndpointToAccessRight(c.Request.Context(), req)
	if err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			msg := validation.FormatValidationErrors(err)
			response.ValidationError(c, exception.ErrValidationError, msg)
			return
		}
		h.log.WithError(err).Error("Failed to link endpoint to access right")
		response.InternalServerError(c, errors.New("could not link endpoint"), "failed to link endpoint")
		return
	}

	response.Success(c, gin.H{"message": "Endpoint linked successfully"})
}



// @Summary      Unlink endpoint from access right
// @Description  Removes an association between an endpoint and a specific access right.
// @Tags         access-rights
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.LinkEndpointRequest true "Unlink Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Endpoint unlinked successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /access-rights/unlink [post]
func (h *AccessController) UnlinkEndpointFromAccessRight(c *gin.Context) {
	var req model.LinkEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	err := h.useCase.UnlinkEndpointFromAccessRight(c.Request.Context(), req)
	if err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			msg := validation.FormatValidationErrors(err)
			response.ValidationError(c, exception.ErrValidationError, msg)
			return
		}
		h.log.WithError(err).Error("Failed to unlink endpoint from access right")
		response.InternalServerError(c, errors.New("could not unlink endpoint"), "failed to unlink endpoint")
		return
	}

	response.Success(c, gin.H{"message": "Endpoint unlinked successfully"})
}


// @Summary      Delete access right
// @Description  Deletes an access right by ID.
// @Tags         access-rights
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Access Right ID"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Access right deleted successfully"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "Access right not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /access-rights/{id} [delete]
func (h *AccessController) DeleteAccessRight(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.useCase.DeleteAccessRight(ctx, id); err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			response.NotFound(c, err, "Access right not found")
		} else {
			response.InternalServerError(c, err, "Failed to delete access right")
		}
		return
	}
	response.Success(c, gin.H{"message": "Access right deleted successfully"})
}

// @Summary      Delete endpoint
// @Description  Deletes an endpoint by ID.
// @Tags         endpoints
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Endpoint ID"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Endpoint deleted successfully"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "Endpoint not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /endpoints/{id} [delete]
func (h *AccessController) DeleteEndpoint(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.useCase.DeleteEndpoint(ctx, id); err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			response.NotFound(c, err, "Endpoint not found")
		} else {
			response.InternalServerError(c, err, "Failed to delete endpoint")
		}
		return
	}
	response.Success(c, gin.H{"message": "Endpoint deleted successfully"})
}

// GetEndpointsDynamic retrieves endpoints based on dynamic filters and sorting via POST request body
// @Summary      Get endpoints with dynamic filters
// @Description  Retrieves a list of endpoints based on dynamic filter and sort criteria provided in the request body.
// @Tags         endpoints
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        filter body querybuilder.DynamicFilter true "Dynamic filter and sort criteria"
// @Success      200  {object}  response.SwaggerEndpointListResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body or filter criteria"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      403  {object}  response.SwaggerErrorResponseWrapper "Forbidden"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /endpoints/search [post]
func (h *AccessController) GetEndpointsDynamic(c *gin.Context) {
	ctx := c.Request.Context()
	var filter querybuilder.DynamicFilter

	if err := c.ShouldBindJSON(&filter); err != nil {
		h.log.WithError(err).Error("failed to bind dynamic filter request body for endpoints")
		response.BadRequest(c, err, "invalid request body for dynamic filter")
		return
	}

	endpoints, total, err := h.useCase.GetEndpointsDynamic(ctx, &filter)
	if err != nil {
		h.log.WithError(err).Error("failed to get endpoints dynamically")
		response.InternalServerError(c, err, "failed to retrieve endpoints")
		return
	}

	response.SuccessResponseWithPaging(c, endpoints, &response.PageMetadata{
		Page:  filter.Page,
		Limit: filter.PageSize,
		Total: total,
	})
}

// GetAccessRightsDynamic retrieves access rights based on dynamic filters and sorting via POST request body
// @Summary      Get access rights with dynamic filters
// @Description  Retrieves a list of access rights based on dynamic filter and sort criteria provided in the request body.
// @Tags         access-rights
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        filter body querybuilder.DynamicFilter true "Dynamic filter and sort criteria"
// @Success      200  {object}  response.SwaggerAccessRightListResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body or filter criteria"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      403  {object}  response.SwaggerErrorResponseWrapper "Forbidden"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /access-rights/search [post]
func (h *AccessController) GetAccessRightsDynamic(c *gin.Context) {
	ctx := c.Request.Context()
	var filter querybuilder.DynamicFilter

	if err := c.ShouldBindJSON(&filter); err != nil {
		h.log.WithError(err).Error("failed to bind dynamic filter request body for access rights")
		response.BadRequest(c, err, "invalid request body for dynamic filter")
		return
	}

	accessRights, total, err := h.useCase.GetAccessRightsDynamic(ctx, &filter)
	if err != nil {
		h.log.WithError(err).Error("failed to get access rights dynamically")
		response.InternalServerError(c, err, "failed to retrieve access rights")
		return
	}

	response.SuccessResponseWithPaging(c, accessRights, &response.PageMetadata{
		Page:  filter.Page,
		Limit: filter.PageSize,
		Total: total,
	})
}
