package tus

import (
	"fmt"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
	tusd "github.com/tus/tusd/v2/pkg/handler"
)

const authenticatedUserIDMetadataKey = "authenticated_user_id"

func BindAuthenticatedMetadata(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {
	userID, ok := authcontext.UserIDFromContext(hook.Context)
	if !ok {
		return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, tusd.NewError(
			"ERR_UNAUTHORIZED_UPLOAD",
			"authenticated user context is required for uploads",
			http.StatusUnauthorized,
		)
	}

	meta := make(tusd.MetaData, len(hook.Upload.MetaData)+1)
	for key, value := range hook.Upload.MetaData {
		meta[key] = value
	}

	meta["user_id"] = userID
	meta[authenticatedUserIDMetadataKey] = userID

	return tusd.HTTPResponse{}, tusd.FileInfoChanges{MetaData: meta}, nil
}

func ValidateUploadMetadata(meta tusd.MetaData, registry *Registry) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {
	uploadType := meta["type"]
	if uploadType == "" {
		return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, tusd.NewError(
			"ERR_UPLOAD_TYPE_REQUIRED",
			"upload type metadata is required",
			http.StatusBadRequest,
		)
	}

	if registry == nil || !registry.Has(uploadType) {
		return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, tusd.NewError(
			"ERR_UNSUPPORTED_UPLOAD_TYPE",
			fmt.Sprintf("unsupported upload type: %s", uploadType),
			http.StatusBadRequest,
		)
	}

	return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, nil
}
