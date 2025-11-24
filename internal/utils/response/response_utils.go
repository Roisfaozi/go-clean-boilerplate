package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, WebResponseSuccess[any]{
		Data: data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, err error, msg string) {
	c.JSON(statusCode, WebResponseError[any]{
		Error:   err.Error(),
		Message: msg,
	})
}

func SuccessResponseWithPaging(c *gin.Context, statusCode int, data interface{}, paging *PageMetadata) {
	c.JSON(statusCode, WebResponseSuccess[any]{
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
