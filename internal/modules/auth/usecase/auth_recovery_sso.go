package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultSSOStateTTL        = 5 * time.Minute
	defaultMaxSSOStates       = 10000
	defaultResendCooldown     = 60 * time.Second
	defaultMaxCooldownEntries = 10000
	verifyEmailFlightKey      = "verify:"
	ssoCallbackFlightKey      = "sso:"
)

type ssoStateContextKey struct{}

// WithSSOState attaches an OAuth2 state value to ctx so HandleSSOCallback can enforce strict one-time consumption.
func WithSSOState(ctx context.Context, state string) context.Context {
	return context.WithValue(ctx, ssoStateContextKey{}, state)
}

// SSOStateFromContext extracts the OAuth2 state attached by WithSSOState.
func SSOStateFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	state, ok := ctx.Value(ssoStateContextKey{}).(string)
	return state, ok
}

type ssoStateStore struct {
	mu      sync.Mutex
	entries map[string]time.Time
}

func newSSOStateStore() *ssoStateStore {
	return &ssoStateStore{entries: make(map[string]time.Time)}
}

func (s *Service) getSSOStateStore() *ssoStateStore {
	s.initMu.Lock()
	defer s.initMu.Unlock()
	if s.ssoStates == nil {
		s.ssoStates = newSSOStateStore()
	}
	return s.ssoStates
}

func ssoRedisKey(providerName, state string) string {
	return "sso:state:" + providerName + ":" + state
}

type cooldownStore struct {
	mu      sync.Mutex
	entries map[string]time.Time
}

func newCooldownStore() *cooldownStore {
	return &cooldownStore{entries: make(map[string]time.Time)}
}

func (s *Service) getResendCooldownStore() *cooldownStore {
	s.initMu.Lock()
	defer s.initMu.Unlock()
	if s.resendCooldowns == nil {
		s.resendCooldowns = newCooldownStore()
	}
	return s.resendCooldowns
}

func (c *cooldownStore) claim(key string, now time.Time, ttl time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, exp := range c.entries {
		if !exp.After(now) {
			delete(c.entries, k)
		}
	}
	if exp, ok := c.entries[key]; ok && exp.After(now) {
		return false
	}
	if len(c.entries) >= defaultMaxCooldownEntries {
		var oldestKey string
		var oldestExp time.Time
		first := true
		for k, exp := range c.entries {
			if first || exp.Before(oldestExp) {
				oldestKey = k
				oldestExp = exp
				first = false
			}
		}
		if !first {
			delete(c.entries, oldestKey)
		}
	}
	c.entries[key] = now.Add(ttl)
	return true
}

func (c *cooldownStore) release(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

func forgotCooldownKey(email string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	sum := sha256.Sum256([]byte(normalized))
	return "auth:cooldown:forgot:" + hex.EncodeToString(sum[:])
}

func verifyCooldownKey(userID string) string {
	return "auth:cooldown:verify:" + userID
}

func (s *Service) claimCooldown(ctx context.Context, key string, cooldown time.Duration) (bool, error) {
	if cooldown <= 0 {
		return true, nil
	}
	if rdb := s.getRedisClient(); rdb != nil {
		acquired, err := rdb.SetNX(ctx, key, "1", cooldown).Result()
		if err != nil {
			if s.log != nil {
				s.log.WithContext(ctx).WithError(err).Warn("Failed to claim resend cooldown in Redis, falling back to in-memory store")
			}
			return s.getResendCooldownStore().claim(key, time.Now(), cooldown), nil
		}
		return acquired, nil
	}
	return s.getResendCooldownStore().claim(key, time.Now(), cooldown), nil
}

func (s *Service) releaseCooldown(ctx context.Context, key string) {
	if rdb := s.getRedisClient(); rdb != nil {
		if err := rdb.Del(ctx, key).Err(); err != nil && s.log != nil {
			s.log.WithContext(ctx).WithError(err).Warn("Failed to release resend cooldown in Redis")
		}
	}
	s.getResendCooldownStore().release(key)
}

func (st *ssoStateStore) record(providerName, state string, now time.Time, ttl time.Duration) {
	key := providerName + ":" + state
	st.mu.Lock()
	defer st.mu.Unlock()

	for k, exp := range st.entries {
		if !exp.After(now) {
			delete(st.entries, k)
		}
	}
	if _, exists := st.entries[key]; !exists && len(st.entries) >= defaultMaxSSOStates {
		var oldestKey string
		var oldestExp time.Time
		first := true
		for k, exp := range st.entries {
			if first || exp.Before(oldestExp) {
				oldestKey = k
				oldestExp = exp
				first = false
			}
		}
		if !first {
			delete(st.entries, oldestKey)
		}
	}
	st.entries[key] = now.Add(ttl)
}

// consume atomically validates and deletes the state entry on first read (strictly one-time use).
func (st *ssoStateStore) consume(providerName, state string, now time.Time) bool {
	key := providerName + ":" + state
	st.mu.Lock()
	defer st.mu.Unlock()

	exp, ok := st.entries[key]
	if !ok {
		return false
	}
	delete(st.entries, key)
	return exp.After(now)
}

type ssoCallbackResult struct {
	loginResponse *model.LoginResponse
	refreshToken  string
}

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("failed to generate random token: %w", err)
	}
	token := hex.EncodeToString(b)

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		s.log.WithContext(ctx).Warnf("Forgot password attempt for non-existent email: %s", email)
		sleepDuration := time.Duration(20+(time.Now().UnixNano()%30)) * time.Millisecond
		time.Sleep(sleepDuration)
		return nil
	}

	cooldown := s.getResendCooldown()
	cooldownKey := forgotCooldownKey(email)
	if cooldown > 0 {
		acquired, _ := s.claimCooldown(ctx, cooldownKey, cooldown)
		if !acquired {
			return nil
		}
	}

	resetToken := &authEntity.PasswordResetToken{
		Email:     email,
		Token:     hashToken(token),
		ExpiresAt: time.Now().Add(15 * time.Minute).UnixMilli(),
	}

	if err := s.tokenRepo.Save(ctx, resetToken); err != nil {
		if cooldown > 0 {
			s.releaseCooldown(ctx, cooldownKey)
		}
		s.log.WithContext(ctx).WithError(err).Error("Failed to save password reset token")
		return nil
	}

	if s.taskDistributor != nil {
		resetURL := strings.TrimRight(s.frontendBaseURL, "/") + "/reset-password?token=" + url.QueryEscape(token)
		taskPayload := &tasks.SendEmailPayload{
			To:      email,
			Subject: "Password Reset Request",
			Body:    fmt.Sprintf("Reset your password using this link: %s. It expires in 15 minutes.", resetURL),
		}
		if err := s.taskDistributor.DistributeTaskSendEmail(ctx, taskPayload); err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to enqueue email task")
		}
	} else {
		s.log.WithContext(ctx).Warnf("Email distributor not configured. Password reset token generated for %s but not logged for security.", email)
	}

	if s.taskDistributor != nil {
		_ = s.taskDistributor.DistributeTaskAuditLog(ctx, auditModel.CreateAuditLogRequest{
			UserID:   user.ID,
			Action:   "FORGOT_PASSWORD_REQUEST",
			Entity:   "User",
			EntityID: user.ID,
		})
	}

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	resetToken, err := s.tokenRepo.FindByToken(ctx, hashToken(token))
	if err != nil {
		return ErrInvalidResetToken
	}

	if time.Now().UnixMilli() > resetToken.ExpiresAt {
		_ = s.tokenRepo.DeleteByEmail(ctx, resetToken.Email)
		return ErrInvalidResetToken
	}

	user, err := s.userRepo.FindByEmail(ctx, resetToken.Email)
	if err != nil {
		return ErrInvalidResetToken
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	err = s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.tokenRepo.RevokeAllSessions(txCtx, user.ID); err != nil {
			s.log.WithContext(txCtx).WithError(err).Error("Failed to revoke sessions during password reset")
			return err
		}

		if err := s.userRepo.Update(txCtx, user); err != nil {
			return err
		}
		return s.tokenRepo.DeleteByEmail(txCtx, resetToken.Email)
	})
	if err != nil {
		return err
	}
	s.evictLoginReplayByUser(user.ID)

	if s.taskDistributor != nil {
		_ = s.taskDistributor.DistributeTaskAuditLog(ctx, auditModel.CreateAuditLogRequest{
			UserID:   user.ID,
			Action:   "PASSWORD_RESET_SUCCESS",
			Entity:   "User",
			EntityID: user.ID,
		})
	}

	return nil
}

func (s *Service) RequestVerification(ctx context.Context, userID string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.EmailVerifiedAt != nil {
		return ErrAlreadyVerified
	}

	cooldown := s.getResendCooldown()
	cooldownKey := verifyCooldownKey(userID)
	if cooldown > 0 {
		acquired, _ := s.claimCooldown(ctx, cooldownKey, cooldown)
		if !acquired {
			return nil
		}
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		if cooldown > 0 {
			s.releaseCooldown(ctx, cooldownKey)
		}
		return fmt.Errorf("failed to generate random token: %w", err)
	}
	token := hex.EncodeToString(b)

	now := time.Now().UnixMilli()
	verificationToken := &authEntity.EmailVerificationToken{
		Email:     user.Email,
		Token:     token,
		ExpiresAt: now + (24 * 60 * 60 * 1000),
		CreatedAt: now,
	}

	if err := s.tokenRepo.SaveVerificationToken(ctx, verificationToken); err != nil {
		if cooldown > 0 {
			s.releaseCooldown(ctx, cooldownKey)
		}
		return err
	}

	if s.taskDistributor != nil {
		verificationURL := strings.TrimRight(s.frontendBaseURL, "/") + "/verify-email?token=" + url.QueryEscape(token)
		taskPayload := &tasks.SendEmailPayload{
			To:      user.Email,
			Subject: "Verify Your Email Address",
			Body:    fmt.Sprintf("Verify your email using this link: %s. It expires in 24 hours.", verificationURL),
		}
		if err := s.taskDistributor.DistributeTaskSendEmail(ctx, taskPayload); err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to enqueue verification email task")
		}
	} else {
		s.log.WithContext(ctx).Warnf("Email distributor not configured. Email verification token generated for %s but not logged for security.", user.Email)
	}

	if s.taskDistributor != nil {
		_ = s.taskDistributor.DistributeTaskAuditLog(ctx, auditModel.CreateAuditLogRequest{
			UserID:   user.ID,
			Action:   "VERIFICATION_EMAIL_REQUESTED",
			Entity:   "User",
			EntityID: user.ID,
		})
	}

	return nil
}

func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	_, err, _ := s.refreshGroup.Do(verifyEmailFlightKey+hashToken(token), func() (interface{}, error) {
		verificationToken, err := s.tokenRepo.FindVerificationToken(ctx, token)
		if err != nil || verificationToken == nil {
			return nil, ErrInvalidVerificationToken
		}

		now := time.Now().UnixMilli()
		if now > verificationToken.ExpiresAt {
			_ = s.tokenRepo.DeleteVerificationTokenByEmail(ctx, verificationToken.Email)
			return nil, ErrInvalidVerificationToken
		}

		user, err := s.userRepo.FindByEmail(ctx, verificationToken.Email)
		if err != nil || user == nil {
			return nil, ErrInvalidVerificationToken
		}

		if user.EmailVerifiedAt != nil {
			_ = s.tokenRepo.DeleteVerificationTokenByEmail(ctx, verificationToken.Email)
			return nil, nil
		}

		user.EmailVerifiedAt = &now

		err = s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
			if err := s.userRepo.Update(txCtx, user); err != nil {
				return err
			}
			return s.tokenRepo.DeleteVerificationTokenByEmail(txCtx, verificationToken.Email)
		})
		if err != nil {
			return nil, err
		}

		if s.taskDistributor != nil {
			_ = s.taskDistributor.DistributeTaskAuditLog(ctx, auditModel.CreateAuditLogRequest{
				UserID:   user.ID,
				Action:   "EMAIL_VERIFIED",
				Entity:   "User",
				EntityID: user.ID,
			})
		}

		return nil, nil
	})

	return err
}

func (s *Service) GetTicket(ctx context.Context, userContext model.UserSessionContext) (string, error) {
	user, err := s.userRepo.FindByID(ctx, userContext.UserID)
	if err != nil {
		s.log.WithContext(ctx).WithError(err).Error("GetTicket failed: could not find user")
		return "", fmt.Errorf("failed to find user: %w", err)
	}
	if user.Status != entity.UserStatusActive {
		return "", ErrAccountSuspended
	}

	return s.ticketManager.CreateTicket(ctx, userContext.UserID, userContext.OrgID, userContext.SessionID, userContext.Role, userContext.Username)
}

func (s *Service) GetSSORedirectURL(ctx context.Context, providerName string, state string) (string, error) {
	provider, exists := s.ssoProviders[providerName]
	if !exists {
		return "", exception.ErrBadRequest
	}

	if state != "" {
		if rdb := s.getRedisClient(); rdb != nil {
			if err := rdb.Set(ctx, ssoRedisKey(providerName, state), "1", defaultSSOStateTTL).Err(); err != nil {
				if s.log != nil {
					s.log.WithContext(ctx).WithError(err).Error("Failed to store SSO state in Redis")
				}
				return "", fmt.Errorf("failed to store SSO state: %w", err)
			}
		} else {
			s.getSSOStateStore().record(providerName, state, time.Now(), defaultSSOStateTTL)
		}
	}

	url := provider.GetLoginURL(state)
	return url, nil
}

// ConsumeSSOState validates and immediately deletes a previously issued SSO state for the given provider.
// Reusing the same state returns exception.ErrUnauthorized.
func (s *Service) ConsumeSSOState(providerName string, state string) error {
	return s.consumeSSOState(context.Background(), providerName, state)
}

func (s *Service) consumeSSOState(ctx context.Context, providerName string, state string) error {
	if providerName == "" || state == "" {
		return exception.ErrUnauthorized
	}
	if rdb := s.getRedisClient(); rdb != nil {
		err := rdb.GetDel(ctx, ssoRedisKey(providerName, state)).Err()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return exception.ErrUnauthorized
			}
			if s.log != nil {
				s.log.WithContext(ctx).WithError(err).Error("Failed to consume SSO state from Redis")
			}
			return fmt.Errorf("failed to consume SSO state: %w", err)
		}
		return nil
	}
	if !s.getSSOStateStore().consume(providerName, state, time.Now()) {
		return exception.ErrUnauthorized
	}
	return nil
}

func (s *Service) HandleSSOCallback(ctx context.Context, providerName string, code string) (*model.LoginResponse, string, error) {
	provider, exists := s.ssoProviders[providerName]
	if !exists {
		return nil, "", exception.ErrBadRequest
	}

	if state, hasState := SSOStateFromContext(ctx); hasState {
		if err := s.consumeSSOState(ctx, providerName, state); err != nil {
			return nil, "", err
		}
	}

	flightKey := ssoCallbackFlightKey + providerName + ":" + hashToken(code)
	result, err, _ := s.refreshGroup.Do(flightKey, func() (interface{}, error) {
		token, err := provider.ExchangeCode(ctx, code)
		if err != nil {
			s.log.Errorf("Failed to exchange code: %v", err)
			return nil, exception.ErrUnauthorized
		}

		userInfo, err := provider.GetUserInfo(ctx, token)
		if err != nil {
			s.log.Errorf("Failed to get user info: %v", err)
			return nil, exception.ErrUnauthorized
		}

		ssoIdentity, err := s.userRepo.FindBySSOIdentity(ctx, providerName, userInfo.ProviderID)

		var usr *entity.User

		if err == nil {
			usr, err = s.userRepo.FindByID(ctx, ssoIdentity.UserID)
			if err != nil {
				s.log.WithContext(ctx).WithError(err).Error("Failed to find user for SSO identity")
				return nil, exception.ErrUnauthorized
			}
		} else {
			usr, err = s.userRepo.FindByEmail(ctx, userInfo.Email)
			if err != nil {
				usrID, _ := uuid.NewV7()
				usr = &entity.User{
					ID:       usrID.String(),
					Email:    userInfo.Email,
					Name:     userInfo.Name,
					Password: "",
					Status:   entity.UserStatusActive,
				}

				err = s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
					if err := s.userRepo.Create(txCtx, usr); err != nil {
						return err
					}

					if s.authz != nil {
						if err := s.authz.AssignDefaultRole(txCtx, usr.ID); err != nil {
							return err
						}
					}

					defaultOrgName := fmt.Sprintf("%s's Workspace", usr.Name)
					slug := pkg.Slugify(fmt.Sprintf("%s-%s", defaultOrgName, usr.ID[:8]))
					defaultOrg := &orgEntity.Organization{
						ID:      uuid.New().String(),
						Name:    defaultOrgName,
						Slug:    slug,
						OwnerID: usr.ID,
						Status:  "active",
					}

					if err := s.orgRepo.Create(txCtx, defaultOrg, "owner"); err != nil {
						return err
					}
					return nil
				})

				if err != nil {
					s.log.WithContext(ctx).WithError(err).Error("Failed to auto-provision user during SSO callback")
					return nil, fmt.Errorf("failed to provision user")
				}
			}

			newSSOIdentity := entity.UserSSOIdentity{
				UserID:     usr.ID,
				Provider:   providerName,
				ProviderID: userInfo.ProviderID,
			}
			if err := s.userRepo.CreateSSOIdentity(ctx, &newSSOIdentity); err != nil {
				s.log.WithContext(ctx).WithError(err).Error("Failed to create SSO identity")
				return nil, fmt.Errorf("failed to link SSO identity")
			}
		}

		if usr.Status != entity.UserStatusActive {
			return nil, ErrAccountSuspended
		}

		var userRole string
		if s.authz != nil {
			roles, err := s.authz.GetRolesForUser(ctx, usr.ID, "")
			if err != nil {
				s.log.WithContext(ctx).WithError(err).Error("Failed to get roles for user during SSO login")
				return nil, fmt.Errorf("failed to get user roles: %w", err)
			}
			if len(roles) > 0 {
				userRole = roles[0]
			}
		}

		accessToken, refreshToken, _, err := s.generateAndStoreTokenPair(ctx, model.UserSessionContext{UserID: usr.ID, Role: userRole, Username: usr.Username})
		if err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to create SSO session")
			return nil, err
		}

		_ = s.tokenRepo.ResetLoginAttempts(ctx, usr.Email)

		if s.taskDistributor != nil {
			_ = s.taskDistributor.DistributeTaskAuditLog(ctx, auditModel.CreateAuditLogRequest{UserID: usr.ID, Action: "SSO_LOGIN", Entity: "User", EntityID: usr.ID})
		}

		s.log.Infof("SSO Login successful for user %s via %s", usr.Email, providerName)

		return &ssoCallbackResult{
			loginResponse: &model.LoginResponse{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				TokenType:    "Bearer",
				ExpiresIn:    int64(s.jwtManager.GetAccessTokenDuration().Seconds()),
				ExpiresAt:    time.Now().Add(s.jwtManager.GetAccessTokenDuration()),
				User:         model.UserInfo{ID: usr.ID, Name: usr.Name, Email: usr.Email, Role: userRole},
			},
			refreshToken: refreshToken,
		}, nil
	})
	if err != nil {
		return nil, "", err
	}

	res, ok := result.(*ssoCallbackResult)
	if !ok || res == nil || res.loginResponse == nil {
		return nil, "", fmt.Errorf("failed to complete SSO callback")
	}

	respCopy := *res.loginResponse
	return &respCopy, res.refreshToken, nil
}
