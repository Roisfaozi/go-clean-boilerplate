package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Service struct {
	jwtManager *jwt.JWTManager
	tokenRepo  repository.TokenRepository
	userRepo   userRepository.UserRepository
	tm         tx.WithTransactionManager
	log        *logrus.Logger
	wsManager  ws.Manager
	Enforcer   permissionUseCase.IEnforcer
	auditUC    auditUseCase.AuditUseCase
}

func NewAuthUsecase(
	jwtManager *jwt.JWTManager,
	tokenRepo repository.TokenRepository,
	userRepo userRepository.UserRepository,
	tm tx.WithTransactionManager,
	log *logrus.Logger,
	wsManager ws.Manager,
	enforcer permissionUseCase.IEnforcer,
	auditUC auditUseCase.AuditUseCase,
) AuthUseCase {
	return &Service{
		jwtManager: jwtManager,
		tokenRepo:  tokenRepo,
		userRepo:   userRepo,
		tm:         tm,
		log:        log,
		wsManager:  wsManager,
		Enforcer:   enforcer,
		auditUC:    auditUC,
	}
}

func (s *Service) generateAndStoreTokenPair(user *entity.User, role, username string) (string, string, string, error) {
	uid, err := uuid.NewV7()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate session id: %w", err)
	}
	sessionID := uid.String()

	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, sessionID, role, username)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate token pair: %w", err)
	}

	session := &model.Auth{
		ID:           sessionID,
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.jwtManager.GetRefreshTokenDuration()),
	}

	if err := s.tokenRepo.StoreToken(context.Background(), session); err != nil {
		s.log.WithError(err).Error("Failed to store session in Redis")
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
		return nil, "", err
	}

	roles, err := s.Enforcer.GetRolesForUser(user.ID)
	if err != nil {
		s.log.WithError(err).Error("Failed to get roles for user during login")
		return nil, "", fmt.Errorf("failed to get user roles: %w", err)
	}
	userRole := ""
	if len(roles) > 0 {
		userRole = roles[0]
	}

	accessToken, refreshToken, sessionID, err := s.generateAndStoreTokenPair(user, userRole, user.Username)
	if err != nil {
		return nil, "", err
	}
	
	// Audit Log: Login (Synchronous)
	if s.auditUC != nil {
		_ = s.auditUC.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			UserID:    user.ID,
			Action:    "LOGIN",
			Entity:    "Auth",
			EntityID:  sessionID,
			IPAddress: request.IPAddress,
			UserAgent: request.UserAgent,
		})
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

	accessTokenDuration := s.jwtManager.GetAccessTokenDuration()
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
		return nil, "", err
	}

	roles, err := s.Enforcer.GetRolesForUser(user.ID)
	if err != nil {
		s.log.WithError(err).Error("Failed to get roles for user during refresh token")
		return nil, "", fmt.Errorf("failed to get user roles: %w", err)
	}
	userRole := ""
	if len(roles) > 0 {
		userRole = roles[0]
	}

	if err := s.RevokeToken(ctx, claims.UserID, claims.SessionID); err != nil {
		s.log.WithError(err).Warn("Failed to revoke old session during refresh")
	}

	newAccessToken, newRefreshToken, _, err := s.generateAndStoreTokenPair(user, userRole, user.Username)
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
	s.log.Infof("Revoking token for user %s with session %s", userID, sessionID)
	
	// Audit Log: Logout (Revoke) (Synchronous)
	if s.auditUC != nil {
		_ = s.auditUC.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			UserID:   userID,
			Action:   "LOGOUT",
			Entity:   "Auth",
			EntityID: sessionID,
		})
	}

	return s.tokenRepo.DeleteToken(ctx, userID, sessionID)
}

func (s *Service) GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error) {
	s.log.Infof("Getting all sessions for user %s", userID)
	return s.tokenRepo.GetUserSessions(ctx, userID)

}

func (s *Service) RevokeAllSessions(ctx context.Context, userID string) error {
	s.log.Infof("Revoking all sessions for user %s", userID)
	
	// Audit Log: Revoke All (Synchronous)
	if s.auditUC != nil {
		_ = s.auditUC.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			UserID:   userID,
			Action:   "REVOKE_ALL_SESSIONS",
			Entity:   "Auth",
			EntityID: userID,
		})
	}

	return s.tokenRepo.RevokeAllSessions(ctx, userID)
}

func (s *Service) GenerateAccessToken(user *entity.User) (string, error) {
	roles, err := s.Enforcer.GetRolesForUser(user.ID)
	if err != nil {
		s.log.WithError(err).Error("Failed to get roles for user when generating access token")
		return "", fmt.Errorf("failed to get user roles: %w", err)
	}
	userRole := ""
	if len(roles) > 0 {
		userRole = roles[0]
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	accessToken, _, err := s.jwtManager.GenerateTokenPair(user.ID, uid.String(), userRole, user.Username)
	return accessToken, err
}

func (s *Service) GenerateRefreshToken(user *entity.User) (string, error) {
	roles, err := s.Enforcer.GetRolesForUser(user.ID)
	if err != nil {
		s.log.WithError(err).Error("Failed to get roles for user when generating refresh token")
		return "", fmt.Errorf("failed to get user roles: %w", err)
	}
	userRole := ""
	if len(roles) > 0 {
		userRole = roles[0]
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	_, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, uid.String(), userRole, user.Username)
	return refreshToken, err
}