package router

import (
	_ "github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
)

// This file contains Swagger annotations for the TUS resumable upload endpoints.
// These are not actual functions but used by swaggo to generate documentation.

// CreateUpload handles upload creation
// @Summary      Create a new resumable upload
// @Description  Initializes a new resumable upload session using the TUS protocol.
// @Tags         upload
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        Tus-Resumable header string true "TUS protocol version (must be 1.0.0)"
// @Param        Upload-Length header int true "Total size of the file in bytes"
// @Param        Upload-Metadata header string true "Base64 encoded metadata (e.g. type, filename)"
// @Header       201 {string} Location "URL to the created upload resource"
// @Success      201 {string} string "Created"
// @Failure      401 {object} response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500 {object} response.SwaggerErrorResponseWrapper "Internal Server Error"
// @Router       /upload/files/ [post]
//
//nolint:unused
func tusCreateUpload() {}

// GetUploadOffset handles checking upload progress
// @Summary      Check upload offset
// @Description  Returns the current offset (number of bytes already uploaded) for a specific resumable upload.
// @Tags         upload
// @Security     BearerAuth
// @Param        id path string true "Upload ID"
// @Param        Tus-Resumable header string true "TUS protocol version (must be 1.0.0)"
// @Header       200 {int} Upload-Offset "Current offset in bytes"
// @Success      200 {string} string "OK"
// @Failure      401 {object} response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      404 {object} response.SwaggerErrorResponseWrapper "Upload not found"
// @Router       /upload/files/{id} [head]
//
//nolint:unused
func tusGetOffset() {}

// UploadChunk handles uploading file data
// @Summary      Upload file chunk
// @Description  Sends a chunk of binary data to an existing resumable upload session.
// @Tags         upload
// @Security     BearerAuth
// @Accept       octet-stream
// @Param        id path string true "Upload ID"
// @Param        Tus-Resumable header string true "TUS protocol version (must be 1.0.0)"
// @Param        Upload-Offset header int true "Current offset of the chunk being sent"
// @Param        Content-Type header string true "Must be application/offset+octet-stream"
// @Success      204 "No Content"
// @Failure      401 {object} response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      409 {object} response.SwaggerErrorResponseWrapper "Overlap or mismatching offset"
// @Router       /upload/files/{id} [patch]
//
//nolint:unused
func tusUploadChunk() {}

// DeleteUpload cancels an upload
// @Summary      Cancel resumable upload
// @Description  Terminates an ongoing resumable upload and removes its partial data.
// @Tags         upload
// @Security     BearerAuth
// @Param        id path string true "Upload ID"
// @Param        Tus-Resumable header string true "TUS protocol version (must be 1.0.0)"
// @Success      204 "No Content"
// @Failure      401 {object} response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      404 {object} response.SwaggerErrorResponseWrapper "Upload not found"
// @Router       /upload/files/{id} [delete]
//
//nolint:unused
func tusDeleteUpload() {}
