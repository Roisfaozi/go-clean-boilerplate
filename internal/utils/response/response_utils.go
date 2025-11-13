package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuccessResponse sends a successful JSON response with the provided data and status code
func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, WebResponse[any]{
		Data: data,
	})
}

// ErrorResponse sends an error JSON response with the provided error message and status code
func ErrorResponse(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, WebResponse[any]{
		Error: err.Error(),
	})
}

// SuccessResponseWithPaging sends a successful JSON response with pagination metadata
func SuccessResponseWithPaging(c *gin.Context, statusCode int, data interface{}, paging *PageMetadata) {
	c.JSON(statusCode, WebResponse[any]{
		Data:   data,
		Paging: paging,
	})
}

// Success creates a 200 OK response with the provided data
func Success(c *gin.Context, data interface{}) {
	SuccessResponse(c, http.StatusOK, data)
}

// Created creates a 201 Created response with the provided data
func Created(c *gin.Context, data interface{}) {
	SuccessResponse(c, http.StatusCreated, data)
}

// BadRequest creates a 400 Bad Request response with the provided error
func BadRequest(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusBadRequest, err)
}

// Unauthorized creates a 401 Unauthorized response with the provided error
func Unauthorized(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusUnauthorized, err)
}

// Forbidden creates a 403 Forbidden response with the provided error
func Forbidden(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusForbidden, err)
}

// NotFound creates a 404 Not Found response with the provided error
func NotFound(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusNotFound, err)
}

// InternalServerError creates a 500 Internal Server Error response with the provided error
func InternalServerError(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusInternalServerError, err)
}

// ValidationError creates a 422 Unprocessable Entity response with the provided error
func ValidationError(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusUnprocessableEntity, err)
}
