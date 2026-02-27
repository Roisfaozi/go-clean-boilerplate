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
	userUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	storageMocks "github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserUseCase_UpdateAvatar_Security(t *testing.T) {
	deps, uc := setupAvatarSecurityTest()
	ctx := context.Background()
	userID := "user-sec-123"

	tests := []struct {
		name        string
		filename    string
		contentType string
		fileContent io.Reader
		errExpected error
	}{
		{
			name:        "Block SVG (potential XSS)",
			filename:    "image.svg",
			contentType: "image/svg+xml",
			fileContent: strings.NewReader(`<?xml version="1.0" standalone="no"?><!DOCTYPE sql SYSTEM "http://malicious.com"><svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"></svg>`),
			errExpected: exception.ErrValidationError,
		},
		{
			name:        "Block HTML disguised as image",
			filename:    "fake.png",
			contentType: "image/png",
			fileContent: strings.NewReader(`<html><body><h1>Not an image</h1><script>alert(1)</script></body></html>`),
			errExpected: exception.ErrValidationError,
		},
		{
			name:        "Block Polyglot (PNG with PHP payload)",
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
			filename:    "exploit.sh",
			contentType: "text/x-shellscript",
			fileContent: strings.NewReader("#!/bin/bash\necho 'hacked'"),
			errExpected: exception.ErrValidationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existingUser := &entity.User{ID: userID, Username: "secuser"}
			deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil).Once()

			if tt.errExpected == nil {
				deps.Storage.On("UploadFile", ctx, mock.Anything, mock.Anything, mock.Anything).Return("http://ok.com", nil).Once()
				deps.Repo.On("Update", ctx, mock.Anything).Return(nil).Once()
				deps.AuditUC.On("LogActivity", ctx, mock.Anything).Return(nil).Once()
			}

			_, err := uc.UpdateAvatar(ctx, userID, tt.fileContent, tt.filename, tt.contentType)

			if tt.errExpected != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.errExpected, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func setupAvatarSecurityTest() (*userTestDeps, userUseCase.UserUseCase) {
	mockEnforcer := new(permMocks.MockIEnforcer)
	deps := &userTestDeps{
		Repo:     new(mocks.MockUserRepository),
		TM:       new(mocking.MockWithTransactionManager),
		Enforcer: mockEnforcer,
		AuditUC:  new(auditMocks.MockAuditUseCase),
		AuthUC:   new(authMocks.MockAuthUseCase),
		Storage:  new(storageMocks.MockProvider),
	}
	log := logrus.New()
	log.SetOutput(io.Discard)
	uc := userUseCase.NewUserUseCase(deps.TM, log, deps.Repo, deps.Enforcer, deps.AuditUC, deps.AuthUC, deps.Storage)
	return deps, uc
}
