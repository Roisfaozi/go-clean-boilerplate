package tus

import (
	_ "github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
)

// This file is strictly for Swagger documentation of the TUS upload endpoint.
// The actual implementation is handled by tusd in router.go.

type TusController struct{}

// TusUpload godoc
// @Summary      Resumable File Upload
// @Description  Implements the TUS protocol for resumable file uploads. Supports POST (initiate), PATCH (content), and HEAD (status).
// @Tags         storage
// @Security     BearerAuth
// @Accept       application/offset+octet-stream
// @Produce      json
// @Param        X-Organization-ID header string true "Organization ID"
// @Param        Upload-Length header int true "Total size of the file in bytes"
// @Param        Tus-Resumable header string true "TUS protocol version (e.g., 1.0.0)"
// @Param        Upload-Metadata header string false "Metadata for the upload (base64 encoded)"
// @Success      201  {string}  string "Created (Location header contains upload URL)"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Router       /upload/files/ [post]
func (ctrl *TusController) TusUpload(c *gin.Context) {}

// TusPatch godoc
// @Summary      Upload File Chunk
// @Description  Uploads a chunk of data to an existing TUS upload.
// @Tags         storage
// @Security     BearerAuth
// @Accept       application/offset+octet-stream
// @Produce      json
// @Param        id   path      string  true  "Upload ID"
// @Param        Upload-Offset header int true "Offset within the file where this chunk starts"
// @Param        Content-Type header string true "application/offset+octet-stream"
// @Success      204  {string}  string "No Content"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Router       /upload/files/{id} [patch]
func (ctrl *TusController) TusPatch(c *gin.Context) {}

// TusHead godoc
// @Summary      Get Upload Status
// @Description  Retrieves the current offset and status of a TUS upload.
// @Tags         storage
// @Security     BearerAuth
// @Param        id   path      string  true  "Upload ID"
// @Success      200  {string}  string "OK (Upload-Offset header contains current progress)"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "Upload not found"
// @Router       /upload/files/{id} [head]
func (ctrl *TusController) TusHead(c *gin.Context) {}
