package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
	"golang.org/x/crypto/bcrypt"
)

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

	resetToken := &authEntity.PasswordResetToken{
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.tokenRepo.Save(ctx, resetToken); err != nil {
		s.log.WithContext(ctx).WithError(err).Error("Failed to save password reset token")
		return nil
	}

	if s.taskDistributor != nil {
		taskPayload := &tasks.SendEmailPayload{
			To:      email,
			Subject: "Password Reset Request",
			Body:    fmt.Sprintf("Your password reset token is: %s. It expires in 15 minutes.", token),
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

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
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
		return err
	}

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

	if s.taskDistributor != nil {
		_ = s.taskDistributor.DistributeTaskAuditLog(ctx, auditModel.CreateAuditLogRequest{
			UserID:   user.ID,
			Action:   "EMAIL_VERIFIED",
			Entity:   "User",
			EntityID: user.ID,
		})
	}

	return nil
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

	url := provider.GetLoginURL(state)
	return url, nil
}

func (s *Service) HandleSSOCallback(ctx context.Context, providerName string, code string) (*model.LoginResponse, string, error) {
	provider, exists := s.ssoProviders[providerName]
	if !exists {
		return nil, "", exception.ErrBadRequest
	}

	token, err := provider.ExchangeCode(ctx, code)
	if err != nil {
		s.log.Errorf("Failed to exchange code: %v", err)
		return nil, "", exception.ErrUnauthorized
	}

	userInfo, err := provider.GetUserInfo(ctx, token)
	if err != nil {
		s.log.Errorf("Failed to get user info: %v", err)
		return nil, "", exception.ErrUnauthorized
	}

	ssoIdentity, err := s.userRepo.FindBySSOIdentity(ctx, providerName, userInfo.ProviderID)

	var usr *entity.User

	if err == nil {
		usr, err = s.userRepo.FindByID(ctx, ssoIdentity.UserID)
		if err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to find user for SSO identity")
			return nil, "", exception.ErrUnauthorized
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
				return nil, "", fmt.Errorf("failed to provision user")
			}
		}

		newSSOIdentity := entity.UserSSOIdentity{
			UserID:     usr.ID,
			Provider:   providerName,
			ProviderID: userInfo.ProviderID,
		}
		if err := s.userRepo.CreateSSOIdentity(ctx, &newSSOIdentity); err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to create SSO identity")
			return nil, "", fmt.Errorf("failed to link SSO identity")
		}
	}

	if usr.Status != entity.UserStatusActive {
		return nil, "", ErrAccountSuspended
	}

	var userRole string
	if s.authz != nil {
		roles, err := s.authz.GetRolesForUser(ctx, usr.ID, "")
		if err != nil {
			s.log.WithContext(ctx).WithError(err).Error("Failed to get roles for user during SSO login")
			return nil, "", fmt.Errorf("failed to get user roles: %w", err)
		}
		if len(roles) > 0 {
			userRole = roles[0]
		}
	}

	accessToken, refreshToken, _, err := s.generateAndStoreTokenPair(ctx, model.UserSessionContext{UserID: usr.ID, Role: userRole, Username: usr.Username})
	if err != nil {
		s.log.WithContext(ctx).WithError(err).Error("Failed to create SSO session")
		return nil, "", err
	}

	_ = s.tokenRepo.ResetLoginAttempts(ctx, usr.Email)

	if s.taskDistributor != nil {
		_ = s.taskDistributor.DistributeTaskAuditLog(ctx, auditModel.CreateAuditLogRequest{UserID: usr.ID, Action: "SSO_LOGIN", Entity: "User", EntityID: usr.ID})
	}

	s.log.Infof("SSO Login successful for user %s via %s", usr.Email, providerName)

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtManager.GetAccessTokenDuration().Seconds()),
		ExpiresAt:    time.Now().Add(s.jwtManager.GetAccessTokenDuration()),
		User:         model.UserInfo{ID: usr.ID, Name: usr.Name, Email: usr.Email, Role: userRole},
	}, refreshToken, nil
}
