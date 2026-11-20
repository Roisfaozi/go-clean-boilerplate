package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/google/uuid"
)

const (
	defaultLoginReplayWindow     = 5 * time.Second
	defaultMaxLoginReplayEntries = 10000
)

type loginReplayEntry struct {
	userID       string
	sessionID    string
	response     *model.LoginResponse
	refreshToken string
	expiresAt    time.Time
}

type loginReplayCache struct {
	mu      sync.Mutex
	entries map[string]loginReplayEntry
}

func newLoginReplayCache() *loginReplayCache {
	return &loginReplayCache{
		entries: make(map[string]loginReplayEntry),
	}
}

func (s *Service) getLoginReplayCache() *loginReplayCache {
	s.initMu.Lock()
	defer s.initMu.Unlock()
	if s.loginReplay == nil {
		s.loginReplay = newLoginReplayCache()
	}
	return s.loginReplay
}

func (c *loginReplayCache) get(key string, now time.Time) (loginReplayEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		return loginReplayEntry{}, false
	}
	if !entry.expiresAt.After(now) {
		delete(c.entries, key)
		return loginReplayEntry{}, false
	}
	return entry, true
}

func (c *loginReplayCache) set(key string, entry loginReplayEntry, now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, existing := range c.entries {
		if !existing.expiresAt.After(now) {
			delete(c.entries, k)
		}
	}
	if _, exists := c.entries[key]; !exists && len(c.entries) >= defaultMaxLoginReplayEntries {
		var oldestKey string
		var oldestExp time.Time
		first := true
		for k, existing := range c.entries {
			if first || existing.expiresAt.Before(oldestExp) {
				oldestKey = k
				oldestExp = existing.expiresAt
				first = false
			}
		}
		if !first {
			delete(c.entries, oldestKey)
		}
	}
	c.entries[key] = entry
}

func (c *loginReplayCache) deleteKey(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

func (c *loginReplayCache) evictByUser(userID string) {
	if userID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, entry := range c.entries {
		if entry.userID == userID {
			delete(c.entries, k)
		}
	}
}

func (c *loginReplayCache) evictBySession(userID, sessionID string) {
	if userID == "" || sessionID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, entry := range c.entries {
		if entry.userID == userID && entry.sessionID == sessionID {
			delete(c.entries, k)
		}
	}
}

func (s *Service) evictLoginReplayByUser(userID string) {
	s.getLoginReplayCache().evictByUser(userID)
}

func (s *Service) evictLoginReplayBySession(userID, sessionID string) {
	s.getLoginReplayCache().evictBySession(userID, sessionID)
}

func normalizeLoginIP(rawIP string) string {
	trimmed := strings.TrimSpace(rawIP)
	if trimmed == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(trimmed); err == nil {
		return strings.TrimSpace(host)
	}
	return trimmed
}

func normalizeLoginUA(rawUA string) string {
	return strings.TrimSpace(rawUA)
}

func loginReplayKey(userID, normalizedIP, normalizedUA string) string {
	sum := sha256.Sum256([]byte(userID + "\x00" + normalizedIP + "\x00" + normalizedUA))
	return hex.EncodeToString(sum[:])
}

func deepCopyLoginResponse(src *model.LoginResponse) *model.LoginResponse {
	if src == nil {
		return nil
	}
	copied := *src
	return &copied
}

type loginReplayResult struct {
	response     *model.LoginResponse
	refreshToken string
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

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
		AccessToken:  hashToken(accessToken),
		RefreshToken: hashToken(refreshToken),
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
	return s.login(ctx, request, true)
}

func (s *Service) login(ctx context.Context, request model.LoginRequest, useReplayCache bool) (*model.LoginResponse, string, error) {
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
				return exception.ErrInternalServer
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

	normalizedIP := normalizeLoginIP(request.IPAddress)
	normalizedUA := normalizeLoginUA(request.UserAgent)
	replayWindow := s.getLoginReplayWindow()

	if !useReplayCache || replayWindow <= 0 || (normalizedIP == "" && normalizedUA == "") {
		resp, refreshToken, _, err := s.issueNewLoginSession(ctx, user, request)
		return resp, refreshToken, err
	}

	key := loginReplayKey(user.ID, normalizedIP, normalizedUA)
	result, err, _ := s.loginReplayGroup.Do(key, func() (interface{}, error) {
		now := time.Now()
		if entry, ok := s.getLoginReplayCache().get(key, now); ok {
			savedSession, getErr := s.tokenRepo.GetToken(ctx, user.ID, entry.sessionID)
			if getErr == nil && savedSession != nil {
				return &loginReplayResult{
					response:     deepCopyLoginResponse(entry.response),
					refreshToken: entry.refreshToken,
				}, nil
			}
			s.getLoginReplayCache().deleteKey(key)
		}

		resp, refreshToken, sessionID, err := s.issueNewLoginSession(ctx, user, request)
		if err != nil {
			return nil, err
		}

		cacheNow := time.Now()
		s.getLoginReplayCache().set(key, loginReplayEntry{
			userID:       user.ID,
			sessionID:    sessionID,
			response:     deepCopyLoginResponse(resp),
			refreshToken: refreshToken,
			expiresAt:    cacheNow.Add(replayWindow),
		}, cacheNow)

		return &loginReplayResult{
			response:     resp,
			refreshToken: refreshToken,
		}, nil
	})
	if err != nil {
		return nil, "", err
	}

	res, ok := result.(*loginReplayResult)
	if !ok || res == nil || res.response == nil {
		return nil, "", fmt.Errorf("failed to complete login")
	}

	return deepCopyLoginResponse(res.response), res.refreshToken, nil
}

func (s *Service) issueNewLoginSession(ctx context.Context, user *entity.User, request model.LoginRequest) (*model.LoginResponse, string, string, error) {
	if s.maxConcurrentSessions > 0 {
		count, err := s.tokenRepo.CountActiveSessions(ctx, user.ID)
		if err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to count user active sessions")
			return nil, "", "", fmt.Errorf("failed to check active sessions: %w", err)
		}
		if count >= s.maxConcurrentSessions {
			return nil, "", "", ErrTooManySessions
		}
	}

	var userRole string
	if s.authz != nil {
		roles, err := s.authz.GetRolesForUser(ctx, user.ID, "")
		if err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to get roles for user during login")
			return nil, "", "", fmt.Errorf("failed to get user roles: %w", err)
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
		return nil, "", "", err
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

	return loginResponse, refreshToken, sessionID, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, string, error) {
	refreshKey := sha256.Sum256([]byte(refreshToken))
	result, err, _ := s.refreshGroup.Do(hex.EncodeToString(refreshKey[:]), func() (interface{}, error) {
		claims, err := s.ValidateRefreshToken(refreshToken)
		if err != nil {
			return nil, err
		}

		user, err := s.userRepo.FindByID(ctx, claims.UserID)
		if err != nil {
			telemetry.UserLoginsTotal.WithLabelValues("failed").Inc()
			return nil, err
		}

		if user.Status != entity.UserStatusActive {
			return nil, ErrAccountSuspended
		}

		var userRole string
		if s.authz != nil {
			roles, err := s.authz.GetRolesForUser(ctx, user.ID, "")
			if err != nil {
				s.log.WithContext(ctx).WithError(err).Error("Failed to get roles for user during refresh token")
				return nil, fmt.Errorf("failed to get user roles: %w", err)
			}
			if len(roles) > 0 {
				userRole = roles[0]
			}
		}

		if err := s.RevokeToken(ctx, claims.UserID, claims.SessionID); err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to revoke old session during refresh")
			return nil, fmt.Errorf("failed to revoke old session during refresh: %w", err)
		}

		newAccessToken, newRefreshToken, _, err := s.generateAndStoreTokenPair(ctx, model.UserSessionContext{
			UserID:   user.ID,
			Role:     userRole,
			Username: user.Username,
		})
		if err != nil {
			return nil, err
		}

		return &refreshResult{
			tokenResponse: &model.TokenResponse{
				AccessToken:  newAccessToken,
				TokenType:    "Bearer",
				RefreshToken: newRefreshToken,
				ExpiresIn:    int64(s.jwtManager.GetAccessTokenDuration().Seconds()),
			},
			refreshToken: newRefreshToken,
		}, nil
	})
	if err != nil {
		return nil, "", err
	}

	res, ok := result.(*refreshResult)
	if !ok || res == nil {
		return nil, "", fmt.Errorf("failed to complete token refresh")
	}

	if res.tokenResponse == nil {
		return nil, "", fmt.Errorf("failed to complete token refresh")
	}

	// Deep copy to prevent concurrent map/struct write issues since the result is shared by singleflight
	tokenResponse := &model.TokenResponse{
		AccessToken:  res.tokenResponse.AccessToken,
		TokenType:    res.tokenResponse.TokenType,
		RefreshToken: res.tokenResponse.RefreshToken,
		ExpiresIn:    res.tokenResponse.ExpiresIn,
	}

	return tokenResponse, res.refreshToken, nil
}

type refreshResult struct {
	tokenResponse *model.TokenResponse
	refreshToken  string
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

	tokenHash := hashToken(tokenString)
	isAccessToken := savedSession.AccessToken == tokenHash
	isRefreshToken := savedSession.RefreshToken == tokenHash
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

	currentSession, err := s.tokenRepo.GetToken(ctx, userID, sessionID)
	if err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			s.evictLoginReplayBySession(userID, sessionID)
			return nil
		}
		return err
	}
	if currentSession == nil {
		s.evictLoginReplayBySession(userID, sessionID)
		return nil
	}

	if s.taskDistributor != nil {
		_ = s.taskDistributor.DistributeTaskAuditLog(ctx, auditModel.CreateAuditLogRequest{
			UserID:   userID,
			Action:   "LOGOUT",
			Entity:   "Auth",
			EntityID: sessionID,
		})
	}

	if err := s.tokenRepo.DeleteToken(ctx, userID, sessionID); err != nil {
		return err
	}
	s.evictLoginReplayBySession(userID, sessionID)
	return nil
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

	if err := s.tokenRepo.RevokeAllSessions(ctx, userID); err != nil {
		return err
	}
	s.evictLoginReplayByUser(userID)
	return nil
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

func (s *Service) GetRefreshTokenDuration() time.Duration {
	return s.jwtManager.GetRefreshTokenDuration()
}
