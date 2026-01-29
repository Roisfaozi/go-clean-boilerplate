package test

import (
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model/converter"
	"github.com/stretchr/testify/assert"
)

func TestConverter_UserToResponse(t *testing.T) {
	now := time.Now().UnixMilli()
	user := &entity.User{
		ID:        "user-123",
		Name:      "Test User",
		Email:     "test@example.com",
		Username:  "testuser",
		AvatarURL: "http://example.com/avatar.png",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	response := converter.UserToResponse(user)

	assert.NotNil(t, response)
	assert.Equal(t, user.ID, response.ID)
	assert.Equal(t, user.Name, response.Name)
	assert.Equal(t, user.Email, response.Email)
	assert.Equal(t, user.Username, response.Username)
	assert.Equal(t, user.AvatarURL, response.AvatarURL)
	assert.Equal(t, user.Status, response.Status)
	assert.Equal(t, user.CreatedAt, response.CreatedAt)
	assert.Equal(t, user.UpdatedAt, response.UpdatedAt)
}

func TestConverter_UserToTokenResponse(t *testing.T) {
	user := &entity.User{
		Token: "some-token-value",
	}

	response := converter.UserToTokenResponse(user)

	assert.NotNil(t, response)
	assert.Equal(t, user.Token, response.Token)
	// Ensure other fields are empty/default
	assert.Empty(t, response.ID)
	assert.Empty(t, response.Name)
}

func TestConverter_UserToEvent(t *testing.T) {
	now := time.Now().UnixMilli()
	user := &entity.User{
		ID:        "event-user",
		Name:      "Event User",
		Email:     "event@example.com",
		Username:  "eventuser",
		CreatedAt: now,
		UpdatedAt: now,
	}

	event := converter.UserToEvent(user)

	assert.NotNil(t, event)
	assert.Equal(t, user.ID, event.ID)
	assert.Equal(t, user.Name, event.Name)
	assert.Equal(t, user.Email, event.Email)
	assert.Equal(t, user.Username, event.Username)
	assert.Equal(t, user.CreatedAt, event.CreatedAt)
	assert.Equal(t, user.UpdatedAt, event.UpdatedAt)
}
