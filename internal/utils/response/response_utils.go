package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, WebResponse[any]{
		Data: data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, WebResponse[any]{
		Error: err.Error(),
	})
}

func SuccessResponseWithPaging(c *gin.Context, statusCode int, data interface{}, paging *PageMetadata) {
	c.JSON(statusCode, WebResponse[any]{
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

func BadRequest(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusBadRequest, err)
}

func Unauthorized(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusUnauthorized, err)
}

func Forbidden(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusForbidden, err)
}

func NotFound(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusNotFound, err)
}

func InternalServerError(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusInternalServerError, err)
}

func ValidationError(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusUnprocessableEntity, err)
}
