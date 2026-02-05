package response

import (
	"errors"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, WebResponseSuccess[any]{
		Data: data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, err error, msg string) {
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	} else {
		errorMsg = http.StatusText(statusCode)
	}

	if statusCode == http.StatusInternalServerError && gin.Mode() == gin.ReleaseMode {
		errorMsg = "Internal Server Error"
	}

	c.JSON(statusCode, WebResponseError[any]{
		Error:   errorMsg,
		Message: msg,
	})
}

func SuccessResponseWithPaging(c *gin.Context, data interface{}, paging *PageMetadata) {
	c.JSON(http.StatusOK, WebResponseSuccess[any]{
		Data:   data,
		Paging: paging,
	})
}

func Success(c *gin.Context, data interface{}) {
	SuccessResponse(c, http.StatusOK, data)
}

func Created(c *gin.Context, data interface{}) {
	SuccessResponse(c, http.StatusCreated, data)
}

func BadRequest(c *gin.Context, err error, msg string) {
	ErrorResponse(c, http.StatusBadRequest, err, msg)
}

func Unauthorized(c *gin.Context, err error, msg string) {
	ErrorResponse(c, http.StatusUnauthorized, err, msg)
}

func Forbidden(c *gin.Context, err error, msg string) {
	ErrorResponse(c, http.StatusForbidden, err, msg)
}

func NotFound(c *gin.Context, err error, msg string) {
	ErrorResponse(c, http.StatusNotFound, err, msg)
}

func InternalServerError(c *gin.Context, err error, msg string) {
	ErrorResponse(c, http.StatusInternalServerError, err, msg)
}

func ValidationError(c *gin.Context, err error, msg string) {
	ErrorResponse(c, http.StatusUnprocessableEntity, err, msg)
}

func Error(c *gin.Context, statusCode int, err error, msg string) {
	// General error handler, useful for custom status codes like 429
	ErrorResponse(c, statusCode, err, msg)
}

func HandleError(c *gin.Context, err error, message string) {
	switch {
	case errors.Is(err, exception.ErrBadRequest):
		BadRequest(c, err, message)
	case errors.Is(err, exception.ErrUnauthorized):
		Unauthorized(c, err, message)
	case errors.Is(err, exception.ErrForbidden):
		Forbidden(c, err, message)
	case errors.Is(err, exception.ErrNotFound):
		NotFound(c, err, message)
	case errors.Is(err, exception.ErrConflict):
		ErrorResponse(c, http.StatusConflict, err, message)
	case errors.Is(err, exception.ErrValidationError), errors.Is(err, exception.ErrUnprocessableEntity):
		ValidationError(c, err, message)
	case errors.Is(err, exception.ErrTooManyRequests):
		Error(c, http.StatusTooManyRequests, err, message)
	default:
		InternalServerError(c, err, message)
	}
}
