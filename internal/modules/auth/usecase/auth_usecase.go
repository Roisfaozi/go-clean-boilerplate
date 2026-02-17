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
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
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
	maxLoginAttempts int
	lockoutDuration  time.Duration
	jwtManager       *jwt.JWTManager
	tokenRepo        repository.TokenRepository
	userRepo         userRepository.UserRepository
	orgRepo          orgRepo.OrganizationRepository
	tm               tx.WithTransactionManager
	log              *logrus.Logger
	wsManager        ws.Manager
	sseManager       *sse.Manager
	Enforcer         permissionUseCase.IEnforcer
	auditUC          auditUseCase.AuditUseCase
	taskDistributor  worker.TaskDistributor
	dummyHash        string
}

func NewAuthUsecase(
	maxLoginAttempts int,
	lockoutDuration time.Duration,
	jwtManager *jwt.JWTManager,
	tokenRepo repository.TokenRepository,
	userRepo userRepository.UserRepository,
	orgRepo orgRepo.OrganizationRepository,
	tm tx.WithTransactionManager,
	log *logrus.Logger,
	wsManager ws.Manager,
	sseManager *sse.Manager,
	enforcer permissionUseCase.IEnforcer,
	auditUC auditUseCase.AuditUseCase,
	taskDistributor worker.TaskDistributor,
) AuthUseCase {
	s := &Service{
		maxLoginAttempts: maxLoginAttempts,
		lockoutDuration:  lockoutDuration,
		jwtManager:       jwtManager,
		tokenRepo:        tokenRepo,
		userRepo:         userRepo,
		orgRepo:          orgRepo,
		tm:               tm,
		log:              log,
		wsManager:        wsManager,
		sseManager:       sseManager,
		Enforcer:         enforcer,
		auditUC:          auditUC,
		taskDistributor:  taskDistributor,
	}

	// Generate dummy hash for timing attack prevention
	// We use the default cost to ensure it matches the real password check duration
	hash, _ := pkg.HashPassword("dummy")
	s.dummyHash = hash

	return s
}

func (s *Service) Register(ctx context.Context, request model.RegisterRequest) (*model.LoginResponse, string, error) {
	// 1. Check if user exists
	if existing, _ := s.userRepo.FindByUsername(ctx, request.Username); existing != nil {
		return nil, "", fmt.Errorf("username already exists")
	}
	if existing, _ := s.userRepo.FindByEmail(ctx, request.Email); existing != nil {
		return nil, "", fmt.Errorf("email already exists")
	}

	// 2. Hash Password
	hashedPassword, err := pkg.HashPassword(request.Password)
	if err != nil {
		return nil, "", err
	}

	userID, _ := uuid.NewV7()
	user := &entity.User{
		ID:       userID.String(),
		Username: request.Username,
		Email:    request.Email,
		Password: hashedPassword,
		Name:     request.Name,
		Status:   entity.UserStatusActive,
	}

	// 3. Transaction: Create User -> Create Default Workspace -> Add Member
	err = s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Create User
		if err := s.userRepo.Create(txCtx, user); err != nil {
			return err
		}

		// Add Default Role (Casbin)
		if s.Enforcer != nil {
			if _, err := s.Enforcer.AddGroupingPolicy(user.ID, "role:user", "global"); err != nil {
				return err
			}
		}

		// Auto-Provisioning: Create Default Workspace
		defaultOrgName := fmt.Sprintf("%s's Workspace", user.Name)
		defaultOrg := &orgEntity.Organization{
			ID:      uuid.New().String(),
			Name:    defaultOrgName,
			Slug:    pkg.Slugify(defaultOrgName + "-" + user.Username), // Simple slug generation
			OwnerID: user.ID,
			Status:  "active",
		}

		// Create Organization (Repo handles adding owner member)
		if err := s.orgRepo.Create(txCtx, defaultOrg, "owner"); err != nil {
			return err
		}

		// Audit Log
		if s.auditUC != nil {
			_ = s.auditUC.LogActivity(txCtx, auditModel.CreateAuditLogRequest{
				UserID:   user.ID,
				Action:   "REGISTER",
				Entity:   "User",
				EntityID: user.ID,
			})
		}
		return nil
	})

	if err != nil {
		return nil, "", err
	}

	telemetry.UserRegistrationsTotal.Inc()

	// 4. Login (Generate Token)
	return s.Login(ctx, model.LoginRequest{
		Username:  request.Username,
		Password:  request.Password,
		IPAddress: request.IPAddress,
		UserAgent: request.UserAgent,
	})
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
	// 1. Check if account is locked
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
			// Timing attack prevention: perform a hash check even if user not found
			pkg.CheckPasswordHash(request.Password, s.dummyHash)
			return ErrInvalidCredentials
		}

		if !pkg.CheckPasswordHash(request.Password, user.Password) {
			// Password Invalid: Handle Lockout Logic
			attempts, incrErr := s.tokenRepo.IncrementLoginAttempts(txCtx, request.Username)
			if incrErr != nil {
				s.log.WithContext(txCtx).WithError(incrErr).Error("Failed to increment login attempts")
			}

			if attempts >= s.maxLoginAttempts {
				if lockErr := s.tokenRepo.LockAccount(txCtx, request.Username, s.lockoutDuration); lockErr != nil {
					s.log.WithContext(txCtx).WithError(lockErr).Error("Failed to lock account")
				}

				// Audit Log: ACCOUNT_LOCKED
				if s.auditUC != nil {
					_ = s.auditUC.LogActivity(txCtx, auditModel.CreateAuditLogRequest{
						UserID:    user.ID, // User ID is known since FindByUsername succeeded
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

		// Password Valid: Reset Attempts
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
	if s.Enforcer != nil {
		roles, err := s.Enforcer.GetRolesForUser(user.ID, "global")
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

	// Broadcast to organization channels
	// 1. Get user organizations
	orgs, err := s.orgRepo.FindUserOrganizations(ctx, user.ID)
	if err != nil {
		// Log but don't fail the request
		s.log.WithContext(ctx).Warnf("Failed to fetch user organizations for notification: %v", err)
	}

	notification := map[string]string{
		"type":    "user_login",
		"user_id": user.ID,
		"message": fmt.Sprintf("User '%s' has just logged in.", user.Name),
		"time":    time.Now().Format(time.RFC3339),
	}
	notificationJSON, _ := json.Marshal(notification)

	if s.wsManager != nil {
		for _, org := range orgs {
			channel := fmt.Sprintf("org_%s_notifications", org.ID)
			s.wsManager.BroadcastToChannel(channel, notificationJSON)
		}
	}

	if s.sseManager != nil {
		// SSE might need similar scoping, but for now we focus on WS
		// TODO: Scope SSE as well if needed
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
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Username:  user.Username,
			Role:      userRole,
			AvatarURL: user.AvatarURL,
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

	if user.Status != entity.UserStatusActive {
		return nil, "", ErrAccountSuspended
	}

	var userRole string
	if s.Enforcer != nil {
		roles, err := s.Enforcer.GetRolesForUser(user.ID, "global")
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
		roles, err := s.Enforcer.GetRolesForUser(user.ID, "global")
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
		roles, err := s.Enforcer.GetRolesForUser(user.ID, "global")
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
	// Generate 32-char hex token unconditionally to prevent timing leaks
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("failed to generate random token: %w", err)
	}
	token := hex.EncodeToString(b)

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Security: Don't reveal if email exists
		s.log.WithContext(ctx).Warnf("Forgot password attempt for non-existent email: %s", email)

		// Simulate DB/Network latency (20-50ms) to prevent user enumeration via timing attacks
		// The success path involves a DB write and task enqueue which typically takes this amount of time.
		// We use a simple modulo on UnixNano to get a pseudo-random duration without importing math/rand.
		sleepDuration := time.Duration(20+(time.Now().UnixNano()%30)) * time.Millisecond
		time.Sleep(sleepDuration)

		return nil
	}

	resetToken := &authEntity.PasswordResetToken{
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.tokenRepo.Save(ctx, resetToken); err != nil {
		// Security: Don't reveal if email exists (DB error)
		s.log.WithContext(ctx).WithError(err).Error("Failed to save password reset token")
		return nil
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
		s.log.WithContext(ctx).Warnf("Email distributor not configured. Password reset token generated for %s but not logged for security.", email)
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	err = s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.Update(txCtx, user); err != nil {
			return err
		}

		// Security: Revoke all existing sessions upon password reset
		if err := s.RevokeAllSessions(txCtx, user.ID); err != nil {
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
		s.log.WithContext(ctx).Warnf("Email distributor not configured. Email verification token generated for %s but not logged for security.", user.Email)
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
