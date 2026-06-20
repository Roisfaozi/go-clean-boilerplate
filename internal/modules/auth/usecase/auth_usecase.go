package usecase

import (
	"context"
	"fmt"
	"time"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sso"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
	publisher        repository.NotificationPublisher
	authz            repository.AuthzManager
	taskDistributor  worker.TaskDistributor
	ticketManager    ws.TicketManager
	ssoProviders     map[string]sso.Provider
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
	publisher repository.NotificationPublisher,
	authz repository.AuthzManager,
	taskDistributor worker.TaskDistributor,
	ticketManager ws.TicketManager,
	ssoProviders map[string]sso.Provider,
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
		publisher:        publisher,
		authz:            authz,
		taskDistributor:  taskDistributor,
		ticketManager:    ticketManager,
		ssoProviders:     ssoProviders,
	}

	hash, _ := pkg.HashPassword("dummy")
	s.dummyHash = hash

	return s
}

func (s *Service) Register(ctx context.Context, request model.RegisterRequest) (*model.LoginResponse, string, error) {
	if existing, _ := s.userRepo.FindByUsername(ctx, request.Username); existing != nil {
		return nil, "", fmt.Errorf("username already exists")
	}
	if existing, _ := s.userRepo.FindByEmail(ctx, request.Email); existing != nil {
		return nil, "", fmt.Errorf("email already exists")
	}

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

	err = s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.Create(txCtx, user); err != nil {
			return err
		}

		if s.authz != nil {
			if err := s.authz.AssignDefaultRole(txCtx, user.ID); err != nil {
				return err
			}
		}

		defaultOrgName := fmt.Sprintf("%s's Workspace", user.Name)
		defaultOrg := &orgEntity.Organization{
			ID:      uuid.New().String(),
			Name:    defaultOrgName,
			Slug:    pkg.Slugify(defaultOrgName + "-" + user.Username),
			OwnerID: user.ID,
			Status:  "active",
		}

		if err := s.orgRepo.Create(txCtx, defaultOrg, "owner"); err != nil {
			return err
		}

		if s.taskDistributor != nil {
			_ = s.taskDistributor.DistributeTaskAuditLog(txCtx, auditModel.CreateAuditLogRequest{
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

	return s.Login(ctx, model.LoginRequest{
		Username:  request.Username,
		Password:  request.Password,
		IPAddress: request.IPAddress,
		UserAgent: request.UserAgent,
	})
}
