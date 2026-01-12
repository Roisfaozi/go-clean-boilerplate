package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	jwtManager      *jwt.JWTManager
	tokenRepo       repository.TokenRepository
	userRepo        userRepository.UserRepository
	tm              tx.WithTransactionManager
	log             *logrus.Logger
	wsManager       ws.Manager
	sseManager      *sse.Manager
	Enforcer        permissionUseCase.IEnforcer
	auditUC         auditUseCase.AuditUseCase
	taskDistributor worker.TaskDistributor
}

func NewAuthUsecase(
	jwtManager *jwt.JWTManager,
	tokenRepo repository.TokenRepository,
	userRepo userRepository.UserRepository,
	tm tx.WithTransactionManager,
	log *logrus.Logger,
	wsManager ws.Manager,
	sseManager *sse.Manager,
	enforcer permissionUseCase.IEnforcer,
	auditUC auditUseCase.AuditUseCase,
	taskDistributor worker.TaskDistributor,
) AuthUseCase {
	return &Service{
		jwtManager:      jwtManager,
		tokenRepo:       tokenRepo,
		userRepo:        userRepo,
		tm:              tm,
		log:             log,
		wsManager:       wsManager,
		sseManager:      sseManager,
		Enforcer:        enforcer,
		auditUC:         auditUC,
		taskDistributor: taskDistributor,
	}
}

func (s *Service) generateAndStoreTokenPair(ctx context.Context, user *entity.User, role, username string) (string, string, string, error) {
	uid, err := uuid.NewV7()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate session id: %w", err)
	}
	sessionID := uid.String()

	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, sessionID, role, username)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate token pair: %w", err)
	}

	now := time.Now()
	session := &model.Auth{
		ID:           sessionID,
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CreatedAt:    now,
		UpdatedAt:    now,
		ExpiresAt:    now.Add(s.jwtManager.GetRefreshTokenDuration()),
	}

	if err := s.tokenRepo.StoreToken(ctx, session); err != nil {
		s.log.WithContext(ctx).WithError(err).Error("Failed to store session in Redis")
		return "", "", "", fmt.Errorf("failed to store session: %w", err)
	}

	return accessToken, refreshToken, sessionID, nil
}

func (s *Service) Login(ctx context.Context, request model.LoginRequest) (*model.LoginResponse, string, error) {
	var user *entity.User
	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		user, err = s.userRepo.FindByUsername(txCtx, request.Username)
		if err != nil {
			return ErrInvalidCredentials
		}

		if !pkg.CheckPasswordHash(request.Password, user.Password) {
			return ErrInvalidCredentials
		}

		return nil
	})

	if err != nil {
		telemetry.UserLoginsTotal.WithLabelValues("failed").Inc()
		return nil, "", err
	}

	var userRole string
	if s.Enforcer != nil {
		roles, err := s.Enforcer.GetRolesForUser(user.ID)
		if err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to get roles for user during login")
			return nil, "", fmt.Errorf("failed to get user roles: %w", err)
		}
		if len(roles) > 0 {
			userRole = roles[0]
		}
	}

	accessToken, refreshToken, sessionID, err := s.generateAndStoreTokenPair(ctx, user, userRole, user.Username)
	if err != nil {
		return nil, "", err
	}

	// Audit Log: Login (Synchronous)
	if s.auditUC != nil {
		if err := s.auditUC.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			UserID:    user.ID,
			Action:    "LOGIN",
			Entity:    "Auth",
			EntityID:  sessionID,
			IPAddress: request.IPAddress,
			UserAgent: request.UserAgent,
		}); err != nil {
			s.log.WithContext(ctx).Warnf("Failed to log activity: %v", err)
		}
	}

	notification := map[string]string{
		"type":    "user_login",
		"user_id": user.ID,
		"message": fmt.Sprintf("User '%s' has just logged in.", user.Name),
		"time":    time.Now().Format(time.RFC3339),
	}
	notificationJSON, _ := json.Marshal(notification)
	if s.wsManager != nil {
		s.wsManager.BroadcastToChannel("global_notifications", notificationJSON)
	}

	if s.sseManager != nil {
		s.sseManager.Broadcast("user_login", notification)
	}

	accessTokenDuration := s.jwtManager.GetAccessTokenDuration()
	telemetry.UserLoginsTotal.WithLabelValues("success").Inc()
	loginResponse := &model.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(accessTokenDuration.Seconds()),
		ExpiresAt:   time.Now().Add(accessTokenDuration),
		User: model.UserInfo{
			ID:       user.ID,
			Name:     user.Name,
			Email:    user.Email,
			Username: user.Username,
			Role:     userRole,
		},
	}

	return loginResponse, refreshToken, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, string, error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, "", err
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		telemetry.UserLoginsTotal.WithLabelValues("failed").Inc()
		return nil, "", err
	}

	var userRole string
	if s.Enforcer != nil {
		roles, err := s.Enforcer.GetRolesForUser(user.ID)
		if err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to get roles for user during refresh token")
			return nil, "", fmt.Errorf("failed to get user roles: %w", err)
		}
		if len(roles) > 0 {
			userRole = roles[0]
		}
	}

	if err := s.RevokeToken(ctx, claims.UserID, claims.SessionID); err != nil {
		s.log.WithContext(ctx).WithError(err).Warn("Failed to revoke old session during refresh")
	}

	newAccessToken, newRefreshToken, _, err := s.generateAndStoreTokenPair(ctx, user, userRole, user.Username)
	if err != nil {
		return nil, "", err
	}

	tokenResponse := &model.TokenResponse{
		AccessToken: newAccessToken,
		TokenType:   "Bearer",
	}

	return tokenResponse, newRefreshToken, nil
}

func (s *Service) ValidateAccessToken(tokenString string) (*jwt.Claims, error) {
	claims, err := s.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return s.validateSession(claims, tokenString)
}

func (s *Service) ValidateRefreshToken(tokenString string) (*jwt.Claims, error) {
	claims, err := s.jwtManager.ValidateRefreshToken(tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return s.validateSession(claims, tokenString)
}

func (s *Service) validateSession(claims *jwt.Claims, tokenString string) (*jwt.Claims, error) {
	savedSession, err := s.tokenRepo.GetToken(context.Background(), claims.UserID, claims.SessionID)
	if err != nil {
		return nil, ErrTokenRevoked
	}

	if savedSession == nil {
		return nil, ErrTokenRevoked
	}

	isAccessToken := savedSession.AccessToken == tokenString
	isRefreshToken := savedSession.RefreshToken == tokenString
	if !isAccessToken && !isRefreshToken {
		return nil, ErrTokenRevoked
	}

	return claims, nil
}

func (s *Service) Verify(ctx context.Context, userID string, sessionID string) (*model.Auth, error) {
	return s.tokenRepo.GetToken(ctx, userID, sessionID)
}

func (s *Service) RevokeToken(ctx context.Context, userID, sessionID string) error {
	s.log.WithContext(ctx).Infof("Revoking token for user %s with session %s", userID, sessionID)

	// Audit Log: Logout (Revoke) (Synchronous)
	if s.auditUC != nil {
		if err := s.auditUC.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			UserID:   userID,
			Action:   "LOGOUT",
			Entity:   "Auth",
			EntityID: sessionID,
		}); err != nil {
			s.log.WithContext(ctx).Warnf("Failed to log activity: %v", err)
		}
	}

	return s.tokenRepo.DeleteToken(ctx, userID, sessionID)
}

func (s *Service) GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error) {
	s.log.WithContext(ctx).Infof("Getting all sessions for user %s", userID)
	return s.tokenRepo.GetUserSessions(ctx, userID)

}

func (s *Service) RevokeAllSessions(ctx context.Context, userID string) error {
	s.log.WithContext(ctx).Infof("Revoking all sessions for user %s", userID)

	// Audit Log: Revoke All (Synchronous)
	if s.auditUC != nil {
		if err := s.auditUC.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			UserID:   userID,
			Action:   "REVOKE_ALL_SESSIONS",
			Entity:   "Auth",
			EntityID: userID,
		}); err != nil {
			s.log.WithContext(ctx).Warnf("Failed to log activity: %v", err)
		}
	}

	return s.tokenRepo.RevokeAllSessions(ctx, userID)
}

func (s *Service) GenerateAccessToken(user *entity.User) (string, error) {
	var userRole string
	if s.Enforcer != nil {
		roles, err := s.Enforcer.GetRolesForUser(user.ID)
		if err != nil {
			s.log.WithContext(context.Background()).WithError(err).Error("Failed to get roles for user when generating access token")
			return "", fmt.Errorf("failed to get user roles: %w", err)
		}
		if len(roles) > 0 {
			userRole = roles[0]
		}
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	accessToken, _, err := s.jwtManager.GenerateTokenPair(user.ID, uid.String(), userRole, user.Username)
	return accessToken, err
}

func (s *Service) GenerateRefreshToken(user *entity.User) (string, error) {
	var userRole string
	if s.Enforcer != nil {
		roles, err := s.Enforcer.GetRolesForUser(user.ID)
		if err != nil {
			s.log.WithContext(context.Background()).WithError(err).Error("Failed to get roles for user when generating refresh token")
			return "", fmt.Errorf("failed to get user roles: %w", err)
		}
		if len(roles) > 0 {
			userRole = roles[0]
		}
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	_, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, uid.String(), userRole, user.Username)
	return refreshToken, err
}

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Security: Don't reveal if email exists
		s.log.WithContext(ctx).Warnf("Forgot password attempt for non-existent email: %s", email)
		return nil
	}

	// Generate 32-char hex token
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("failed to generate random token: %w", err)
	}
	token := hex.EncodeToString(b)

	resetToken := &authEntity.PasswordResetToken{
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.tokenRepo.Save(ctx, resetToken); err != nil {
		return err
	}

	// Send Email Async
	if s.taskDistributor != nil {
		taskPayload := &tasks.SendEmailPayload{
			To:      email,
			Subject: "Password Reset Request",
			Body:    fmt.Sprintf("Your password reset token is: %s. It expires in 15 minutes.", token),
		}
		if err := s.taskDistributor.DistributeTaskSendEmail(ctx, taskPayload); err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to enqueue email task")
			// We log the error but don't fail the request, allowing manual fallback if needed
		}
	} else {
		// Fallback logging if distributor is not configured
		s.log.WithContext(ctx).Infof("PASSWORD RESET TOKEN for %s: %s (Expires in 15m)", email, token)
	}

	if s.auditUC != nil {
		if err := s.auditUC.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			UserID:   user.ID,
			Action:   "FORGOT_PASSWORD_REQUEST",
			Entity:   "User",
			EntityID: user.ID,
		}); err != nil {
			s.log.WithContext(ctx).Warnf("Failed to log activity: %v", err)
		}
	}

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	resetToken, err := s.tokenRepo.FindByToken(ctx, token)
	if err != nil {
		return ErrInvalidResetToken
	}

	if time.Now().After(resetToken.ExpiresAt) {
		_ = s.tokenRepo.DeleteByEmail(ctx, resetToken.Email)
		return ErrInvalidResetToken
	}

	user, err := s.userRepo.FindByEmail(ctx, resetToken.Email)
	if err != nil {
		return ErrInvalidResetToken
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	err = s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.Update(txCtx, user); err != nil {
			return err
		}
		return s.tokenRepo.DeleteByEmail(txCtx, resetToken.Email)
	})

	if err != nil {
		return err
	}

	if s.auditUC != nil {
		if err := s.auditUC.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			UserID:   user.ID,
			Action:   "PASSWORD_RESET_SUCCESS",
			Entity:   "User",
			EntityID: user.ID,
		}); err != nil {
			s.log.WithContext(ctx).Warnf("Failed to log activity: %v", err)
		}
	}

	return nil
}

func (s *Service) RequestVerification(ctx context.Context, userID string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if already verified
	if user.EmailVerifiedAt != nil {
		return ErrAlreadyVerified
	}

	// Generate 32-char hex token
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("failed to generate random token: %w", err)
	}
	token := hex.EncodeToString(b)

	now := time.Now().UnixMilli()
	verificationToken := &authEntity.EmailVerificationToken{
		Email:     user.Email,
		Token:     token,
		ExpiresAt: now + (24 * 60 * 60 * 1000), // 24 hours in milliseconds
		CreatedAt: now,
	}

	if err := s.tokenRepo.SaveVerificationToken(ctx, verificationToken); err != nil {
		return err
	}

	// Send Email Async
	if s.taskDistributor != nil {
		taskPayload := &tasks.SendEmailPayload{
			To:      user.Email,
			Subject: "Verify Your Email Address",
			Body:    fmt.Sprintf("Please verify your email by using this token: %s. It expires in 24 hours.", token),
		}
		if err := s.taskDistributor.DistributeTaskSendEmail(ctx, taskPayload); err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to enqueue verification email task")
		}
	} else {
		s.log.WithContext(ctx).Infof("EMAIL VERIFICATION TOKEN for %s: %s (Expires in 24h)", user.Email, token)
	}

	if s.auditUC != nil {
		if err := s.auditUC.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			UserID:   user.ID,
			Action:   "VERIFICATION_EMAIL_REQUESTED",
			Entity:   "User",
			EntityID: user.ID,
		}); err != nil {
			s.log.WithContext(ctx).Warnf("Failed to log activity: %v", err)
		}
	}

	return nil
}

func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	verificationToken, err := s.tokenRepo.FindVerificationToken(ctx, token)
	if err != nil {
		return ErrInvalidVerificationToken
	}

	now := time.Now().UnixMilli()
	if now > verificationToken.ExpiresAt {
		_ = s.tokenRepo.DeleteVerificationTokenByEmail(ctx, verificationToken.Email)
		return ErrInvalidVerificationToken
	}

	user, err := s.userRepo.FindByEmail(ctx, verificationToken.Email)
	if err != nil {
		return ErrInvalidVerificationToken
	}

	// Already verified check
	if user.EmailVerifiedAt != nil {
		_ = s.tokenRepo.DeleteVerificationTokenByEmail(ctx, verificationToken.Email)
		return ErrAlreadyVerified
	}

	user.EmailVerifiedAt = &now

	err = s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.Update(txCtx, user); err != nil {
			return err
		}
		return s.tokenRepo.DeleteVerificationTokenByEmail(txCtx, verificationToken.Email)
	})

	if err != nil {
		return err
	}

	if s.auditUC != nil {
		if err := s.auditUC.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			UserID:   user.ID,
			Action:   "EMAIL_VERIFIED",
			Entity:   "User",
			EntityID: user.ID,
		}); err != nil {
			s.log.WithContext(ctx).Warnf("Failed to log activity: %v", err)
		}
	}

	return nil
}
