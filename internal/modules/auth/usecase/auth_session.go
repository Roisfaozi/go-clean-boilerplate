package usecase

import (
	"context"
	"fmt"
	"time"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/google/uuid"
)

func (s *Service) generateAndStoreTokenPair(ctx context.Context, userContext model.UserSessionContext) (string, string, string, error) {
	uid, err := uuid.NewV7()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate session id: %w", err)
	}
	sessionID := uid.String()

	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(jwt.UserContext{
		UserID:    userContext.UserID,
		SessionID: sessionID,
		Role:      userContext.Role,
		Username:  userContext.Username,
		OrgID:     userContext.OrgID,
	})
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate token pair: %w", err)
	}

	now := time.Now()
	session := &model.Auth{
		ID:           sessionID,
		UserID:       userContext.UserID,
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
	locked, ttl, err := s.tokenRepo.IsAccountLocked(ctx, request.Username)
	if err != nil {
		s.log.WithContext(ctx).WithError(err).Error("Failed to check account lock status")
		return nil, "", fmt.Errorf("failed to check account status")
	}
	if locked {
		return nil, "", fmt.Errorf("%w: try again in %v", ErrAccountLocked, ttl.Round(time.Second))
	}

	var user *entity.User
	err = s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		user, err = s.userRepo.FindByUsername(txCtx, request.Username)
		if err != nil {
			pkg.CheckPasswordHash(request.Password, s.dummyHash)
			return ErrInvalidCredentials
		}

		if !pkg.CheckPasswordHash(request.Password, user.Password) {
			attempts, incrErr := s.tokenRepo.IncrementLoginAttempts(txCtx, request.Username)
			if incrErr != nil {
				s.log.WithContext(txCtx).WithError(incrErr).Error("Failed to increment login attempts")
			}

			if attempts >= s.maxLoginAttempts {
				if lockErr := s.tokenRepo.LockAccount(txCtx, request.Username, s.lockoutDuration); lockErr != nil {
					s.log.WithContext(txCtx).WithError(lockErr).Error("Failed to lock account")
				}

				if s.taskDistributor != nil {
					_ = s.taskDistributor.DistributeTaskAuditLog(txCtx, auditModel.CreateAuditLogRequest{
						UserID:    user.ID,
						Action:    "ACCOUNT_LOCKED",
						Entity:    "User",
						EntityID:  user.ID,
						IPAddress: request.IPAddress,
						UserAgent: request.UserAgent,
					})
				}
				return fmt.Errorf("%w: too many failed attempts", ErrAccountLocked)
			}

			return ErrInvalidCredentials
		}

		if resetErr := s.tokenRepo.ResetLoginAttempts(txCtx, request.Username); resetErr != nil {
			s.log.WithContext(txCtx).WithError(resetErr).Error("Failed to reset login attempts")
		}

		if user.Status != entity.UserStatusActive {
			return ErrAccountSuspended
		}

		return nil
	})

	if err != nil {
		telemetry.UserLoginsTotal.WithLabelValues("failed").Inc()
		return nil, "", err
	}

	var userRole string
	if s.authz != nil {
		roles, err := s.authz.GetRolesForUser(ctx, user.ID, "")
		if err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to get roles for user during login")
			return nil, "", fmt.Errorf("failed to get user roles: %w", err)
		}
		if len(roles) > 0 {
			userRole = roles[0]
		}
	}

	accessToken, refreshToken, sessionID, err := s.generateAndStoreTokenPair(ctx, model.UserSessionContext{
		UserID:   user.ID,
		Role:     userRole,
		Username: user.Username,
	})
	if err != nil {
		return nil, "", err
	}

	if s.taskDistributor != nil {
		_ = s.taskDistributor.DistributeTaskAuditLog(ctx, auditModel.CreateAuditLogRequest{
			UserID:    user.ID,
			Action:    "LOGIN",
			Entity:    "Auth",
			EntityID:  sessionID,
			IPAddress: request.IPAddress,
			UserAgent: request.UserAgent,
		})
	}

	orgs, err := s.orgRepo.FindUserOrganizations(ctx, user.ID)
	if err != nil {
		s.log.WithContext(ctx).Warnf("Failed to fetch user organizations for notification: %v", err)
	}

	var orgIDs []string
	for _, org := range orgs {
		orgIDs = append(orgIDs, org.ID)
	}

	userInfo := model.UserInfo{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Username:  user.Username,
		Role:      userRole,
		AvatarURL: user.AvatarURL,
	}

	if s.publisher != nil {
		s.publisher.PublishUserLoggedIn(ctx, userInfo, orgIDs)
	}

	accessTokenDuration := s.jwtManager.GetAccessTokenDuration()
	telemetry.UserLoginsTotal.WithLabelValues("success").Inc()
	loginResponse := &model.LoginResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessTokenDuration.Seconds()),
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(accessTokenDuration),
		User:         userInfo,
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

	if user.Status != entity.UserStatusActive {
		return nil, "", ErrAccountSuspended
	}

	var userRole string
	if s.authz != nil {
		roles, err := s.authz.GetRolesForUser(ctx, user.ID, "")
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

	newAccessToken, newRefreshToken, _, err := s.generateAndStoreTokenPair(ctx, model.UserSessionContext{
		UserID:   user.ID,
		Role:     userRole,
		Username: user.Username,
	})
	if err != nil {
		return nil, "", err
	}

	tokenResponse := &model.TokenResponse{
		AccessToken:  newAccessToken,
		TokenType:    "Bearer",
		RefreshToken: newRefreshToken,
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

	if s.taskDistributor != nil {
		_ = s.taskDistributor.DistributeTaskAuditLog(ctx, auditModel.CreateAuditLogRequest{
			UserID:   userID,
			Action:   "LOGOUT",
			Entity:   "Auth",
			EntityID: sessionID,
		})
	}

	return s.tokenRepo.DeleteToken(ctx, userID, sessionID)
}

func (s *Service) GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error) {
	s.log.WithContext(ctx).Infof("Getting all sessions for user %s", userID)
	return s.tokenRepo.GetUserSessions(ctx, userID)
}

func (s *Service) RevokeAllSessions(ctx context.Context, userID string) error {
	s.log.WithContext(ctx).Infof("Revoking all sessions for user %s", userID)

	if s.taskDistributor != nil {
		_ = s.taskDistributor.DistributeTaskAuditLog(ctx, auditModel.CreateAuditLogRequest{
			UserID:   userID,
			Action:   "REVOKE_ALL_SESSIONS",
			Entity:   "Auth",
			EntityID: userID,
		})
	}

	return s.tokenRepo.RevokeAllSessions(ctx, userID)
}

func (s *Service) GenerateAccessToken(user *entity.User) (string, error) {
	var userRole string
	if s.authz != nil {
		roles, err := s.authz.GetRolesForUser(context.Background(), user.ID, "")
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
	accessToken, _, err := s.jwtManager.GenerateTokenPair(jwt.UserContext{
		UserID:    user.ID,
		SessionID: uid.String(),
		Role:      userRole,
		Username:  user.Username,
	})
	return accessToken, err
}

func (s *Service) GenerateRefreshToken(user *entity.User) (string, error) {
	var userRole string
	if s.authz != nil {
		roles, err := s.authz.GetRolesForUser(context.Background(), user.ID, "")
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
	_, refreshToken, err := s.jwtManager.GenerateTokenPair(jwt.UserContext{
		UserID:    user.ID,
		SessionID: uid.String(),
		Role:      userRole,
		Username:  user.Username,
	})
	return refreshToken, err
}
