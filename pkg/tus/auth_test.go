package tus

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tusd "github.com/tus/tusd/v2/pkg/handler"
)

func TestBindAuthenticatedMetadata_Positive_OverridesClientUserID(t *testing.T) {
	ctx := authcontext.WithUserID(context.Background(), "user-123")

	_, changes, err := BindAuthenticatedMetadata(tusd.HookEvent{
		Context: ctx,
		Upload: tusd.FileInfo{
			MetaData: tusd.MetaData{
				"type":    "avatar",
				"user_id": "victim-user",
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "user-123", changes.MetaData["user_id"])
	assert.Equal(t, "user-123", changes.MetaData["authenticated_user_id"])
	assert.Equal(t, "avatar", changes.MetaData["type"])
}

func TestBindAuthenticatedMetadata_Negative_RejectsMissingUserContext(t *testing.T) {
	_, _, err := BindAuthenticatedMetadata(tusd.HookEvent{
		Context: context.Background(),
		Upload: tusd.FileInfo{
			MetaData: tusd.MetaData{"type": "avatar"},
		},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_UNAUTHORIZED_UPLOAD")
}

func TestValidateUploadMetadata_Positive_AllowsRegisteredType(t *testing.T) {
	registry := NewRegistry()
	registry.Register("avatar", &MockHook{})

	_, _, err := ValidateUploadMetadata(tusd.MetaData{"type": "avatar"}, registry)

	require.NoError(t, err)
}

func TestValidateUploadMetadata_Negative_RejectsMissingType(t *testing.T) {
	registry := NewRegistry()
	registry.Register("avatar", &MockHook{})

	_, _, err := ValidateUploadMetadata(tusd.MetaData{}, registry)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_UPLOAD_TYPE_REQUIRED")
}

func TestValidateUploadMetadata_Negative_RejectsUnknownType(t *testing.T) {
	registry := NewRegistry()
	registry.Register("avatar", &MockHook{})

	_, _, err := ValidateUploadMetadata(tusd.MetaData{"type": "../../etc/passwd"}, registry)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_UNSUPPORTED_UPLOAD_TYPE")
}
