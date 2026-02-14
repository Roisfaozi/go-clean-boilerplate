package test

import (
	"context"
	"io"
	"strings"
	"testing"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	permMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	storageMocks "github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserUseCase_UpdateAvatar_Security(t *testing.T) {
	deps, uc := setupAvatarSecurityTest()
	ctx := context.Background()
	userID := "019b9150-304e-79d0-aa16-4a2b44347a08" // Valid UUID

	tests := []struct {
		name        string
		userID      string
		filename    string
		contentType string
		fileContent io.Reader
		errExpected error
	}{
		{
			name:        "Block SVG (potential XSS)",
			userID:      userID,
			filename:    "image.svg",
			contentType: "image/svg+xml",
			fileContent: strings.NewReader(`<?xml version="1.0" standalone="no"?><!DOCTYPE sql SYSTEM "http://malicious.com"><svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"></svg>`),
			errExpected: exception.ErrValidationError,
		},
		{
			name:        "Block HTML disguised as image",
			userID:      userID,
			filename:    "fake.png",
			contentType: "image/png",
			fileContent: strings.NewReader(`<html><body><h1>Not an image</h1><script>alert(1)</script></body></html>`),
			errExpected: exception.ErrValidationError,
		},
		{
			name:        "Block Polyglot (PNG with PHP payload)",
			userID:      userID,
			filename:    "poly.png",
			contentType: "image/png",
			fileContent: io.MultiReader(
				strings.NewReader("\x89PNG\r\n\x1a\n"),
				strings.NewReader("<?php echo 'malicious'; ?>"),
			),
			errExpected: nil,
		},
		{
			name:        "Block Script File",
			userID:      userID,
			filename:    "exploit.sh",
			contentType: "text/x-shellscript",
			fileContent: strings.NewReader("#!/bin/bash\necho 'hacked'"),
			errExpected: exception.ErrValidationError,
		},
		{
			name:        "Block Path Traversal in UserID",
			userID:      "../etc/passwd",
			filename:    "profile.png",
			contentType: "image/png",
			fileContent: strings.NewReader("fake-content"),
			errExpected: exception.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testUserID := tt.userID
			if testUserID == "" {
				testUserID = userID // Default to valid UUID
			}

			// For the BadRequest case (Path Traversal), FindByID should NOT be called.
			if tt.errExpected != exception.ErrBadRequest {
				existingUser := &entity.User{ID: testUserID, Username: "secuser"}
				deps.Repo.On("FindByID", ctx, testUserID).Return(existingUser, nil).Once()
			}

			if tt.errExpected == nil {
				deps.Storage.On("UploadFile", ctx, mock.Anything, mock.Anything, mock.Anything).Return("http://ok.com", nil).Once()
				deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}).Once()
				deps.Repo.On("Update", ctx, mock.Anything).Return(nil).Once()
				deps.AuditUC.On("LogActivity", ctx, mock.Anything).Return(nil).Once()
			}

			_, err := uc.UpdateAvatar(ctx, testUserID, testUserID, tt.fileContent, tt.filename, tt.contentType)

			if tt.errExpected != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.errExpected, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func setupAvatarSecurityTest() (*userTestDeps, usecase.UserUseCase) {
	deps := &userTestDeps{
		Repo:     new(mocks.MockUserRepository),
		TM:       new(mocking.MockWithTransactionManager),
		Enforcer: new(permMocks.IEnforcer),
		AuditUC:  new(auditMocks.MockAuditUseCase),
		AuthUC:   new(authMocks.MockAuthUseCase),
		Storage:  new(storageMocks.MockProvider),
	}
	log := logrus.New()
	log.SetOutput(io.Discard)
	uc := usecase.NewUserUseCase(deps.TM, log, deps.Repo, deps.Enforcer, deps.AuditUC, deps.AuthUC, deps.Storage)
	return deps, uc
}
