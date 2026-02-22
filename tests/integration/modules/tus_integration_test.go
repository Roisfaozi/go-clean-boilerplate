//go:build integration
// +build integration

package modules

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tus/tusd/v2/pkg/handler"

	auditRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	userUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tus"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
)

const (
	rustfsBucket = "test-bucket"
	rustfsRegion = "us-east-1"
)

func setupTUSDeps(t *testing.T, env *setup.TestEnvironment, s3Client *s3.Client, s3URL string) (*handler.Handler, userUseCase.UserUseCase, string) {
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger, nil)
	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	jwtManager := jwt.NewJWTManager("test-secret", "test-refresh", time.Hour, time.Hour*24)
	orgRepo := orgRepository.NewOrganizationRepository(env.DB)
	authUC := authUseCase.NewAuthUsecase(5, 30*time.Minute, jwtManager, tokenRepo, userRepo, orgRepo, tm, env.Logger, nil, nil, env.Enforcer, auditUC, nil, nil)

	uc := userUseCase.NewUserUseCase(tm, env.Logger, userRepo, env.Enforcer, auditUC, authUC, nil)

	registry := tus.NewRegistry()
	avatarHook := &userUseCase.AvatarHook{UserUseCase: uc}
	registry.Register("avatar", avatarHook)

	handler, err := tus.NewHandler(tus.Config{
		StorageDriver: "s3",
		S3Bucket:      rustfsBucket,
		S3Endpoint:    s3URL,
		BasePath:      "/files/",
	}, registry, s3Client, env.Logger)
	require.NoError(t, err)

	testUser := setup.CreateTestUser(t, env.DB, "tususer", "tus@example.com", "password123")
	return handler, uc, testUser.ID
}

func TestTUS_Integration_Lifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	ctx := context.Background()

	s3URL, s3Bucket := setup.SetupRustFS(t, ctx)

	awsCfg, _ := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(rustfsRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("rustfsadmin", "rustfsadmin", "")),
	)
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s3URL)
		o.UsePathStyle = true
	})

	// Wait and create bucket
	time.Sleep(2 * time.Second)
	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s3Bucket),
	})
	require.NoError(t, err)

	tusHandler, _, userID := setupTUSDeps(t, env, s3Client, s3URL)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/files/", gin.WrapH(http.StripPrefix("/files/", tusHandler)))
	router.PATCH("/files/*any", gin.WrapH(http.StripPrefix("/files/", tusHandler)))
	router.HEAD("/files/*any", gin.WrapH(http.StripPrefix("/files/", tusHandler)))

	t.Run("Create Upload", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/files/", nil)
		req.Header.Set("Tus-Resumable", "1.0.0")
		req.Header.Set("Upload-Length", "5")

		userIDBase64 := base64.StdEncoding.EncodeToString([]byte(userID))
		req.Header.Set("Upload-Metadata", fmt.Sprintf("filename dGVzdC50eHQ=,type YXZhdGFy,user_id %s", userIDBase64))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		location := w.Header().Get("Location")
		require.NotEmpty(t, location)

		// Location matches /files/{id} (or absolute URL)
		parts := strings.Split(location, "/")
		uploadID := parts[len(parts)-1]
		require.NotEmpty(t, uploadID)

		t.Run("Upload Chunk", func(t *testing.T) {
			body := []byte("hello")
			req, _ := http.NewRequest("PATCH", location, bytes.NewReader(body))
			req.Header.Set("Tus-Resumable", "1.0.0")
			req.Header.Set("Upload-Offset", "0")
			req.Header.Set("Content-Type", "application/offset+octet-stream")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, http.StatusNoContent, w.Code)
			offset := w.Header().Get("Upload-Offset")
			assert.Equal(t, "5", offset)

			time.Sleep(2 * time.Second)

			var user struct{ AvatarURL string }
			err := env.DB.Table("users").Select("avatar_url").Where("id = ?", userID).Scan(&user).Error
			require.NoError(t, err)

			assert.Contains(t, user.AvatarURL, rustfsBucket)
			assert.Contains(t, user.AvatarURL, uploadID)
		})

		t.Run("Invalid Chunk Offset", func(t *testing.T) {
			body := []byte("world")
			req, _ := http.NewRequest("PATCH", location, bytes.NewReader(body))
			req.Header.Set("Tus-Resumable", "1.0.0")
			req.Header.Set("Upload-Offset", "0")
			req.Header.Set("Content-Type", "application/offset+octet-stream")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusConflict, w.Code)
		})
	})
}
