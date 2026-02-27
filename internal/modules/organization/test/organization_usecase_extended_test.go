package test

import (
	"context"
	"errors"
	"io"
	"testing"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	orgMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	permMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type orgTestDeps struct {
	Repo            *orgMocks.MockOrganizationRepository
	MemberRepo      *orgMocks.MockOrganizationMemberRepository
	InviteRepo      *orgMocks.MockInvitationRepository
	UserRepo        *userMocks.MockUserRepository
	TM              *mocking.MockWithTransactionManager
	Enforcer        *permMocks.IEnforcer
	TaskDistributor *mocking.MockTaskDistributor
	PresenceReader  *authMocks.MockPresenceReader
}

func setupOrganizationUseCase(deps *orgTestDeps) usecase.OrganizationUseCase {
	log := logrus.New()
	log.SetOutput(io.Discard)
	var enf permissionUseCase.IEnforcer = deps.Enforcer
	return usecase.NewOrganizationUseCase(log, deps.TM, deps.Repo, deps.MemberRepo, enf)
}

func setupMemberUseCase(deps *orgTestDeps) usecase.OrganizationMemberUseCase {
	log := logrus.New()
	log.SetOutput(io.Discard)
	var enf permissionUseCase.IEnforcer = deps.Enforcer
	return usecase.NewOrganizationMemberUseCase(log, deps.TM, deps.MemberRepo, deps.Repo, deps.InviteRepo, deps.UserRepo, deps.TaskDistributor, enf, deps.PresenceReader, "http://frontend")
}

func TestCreateOrganization_Extended(t *testing.T) {
	deps := &orgTestDeps{
		Repo:       new(orgMocks.MockOrganizationRepository),
		MemberRepo: new(orgMocks.MockOrganizationMemberRepository),
		TM:         new(mocking.MockWithTransactionManager),
		Enforcer:   new(permMocks.IEnforcer),
	}
	uc := setupOrganizationUseCase(deps)

	req := model.CreateOrganizationRequest{Name: "New Org", Slug: "new-org"}
	userID := "user1"

	t.Run("Transaction_Failure", func(t *testing.T) {
		deps.Repo.On("FindBySlug", mock.Anything, "new-org").Return(nil, nil).Once()
		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(exception.ErrInternalServer)

		deps.Repo.On("Create", mock.Anything, mock.Anything, "owner").Return(errors.New("db error")).Once()

		_, err := uc.Create(context.Background(), userID, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})
}

func TestInviteMember_Extended(t *testing.T) {
	deps := &orgTestDeps{
		Repo:            new(orgMocks.MockOrganizationRepository),
		MemberRepo:      new(orgMocks.MockOrganizationMemberRepository),
		InviteRepo:      new(orgMocks.MockInvitationRepository),
		UserRepo:        new(userMocks.MockUserRepository),
		TM:              new(mocking.MockWithTransactionManager),
		Enforcer:        new(permMocks.IEnforcer),
		TaskDistributor: new(mocking.MockTaskDistributor),
	}
	uc := setupMemberUseCase(deps)

	orgID := "org1"
	req := model.InviteMemberRequest{Email: "invitee@example.com", Role: "member"}
	actorID := "admin1"

	t.Run("Email_Send_Failure", func(t *testing.T) {
		// Mock checks
		deps.MemberRepo.On("FindMember", mock.Anything, orgID, actorID).Return(&entity.OrganizationMember{Role: "admin"}, nil).Once()
		deps.Repo.On("FindByID", mock.Anything, orgID).Return(&entity.Organization{ID: orgID, Name: "Test Org"}, nil).Once()
		deps.MemberRepo.On("IsMemberByEmail", mock.Anything, orgID, req.Email).Return(false, nil).Once()
		deps.InviteRepo.On("FindByEmail", mock.Anything, orgID, req.Email).Return(nil, nil).Once()

		// Transaction
		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(nil)

		deps.InviteRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		// Email failure
		deps.TaskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.MatchedBy(func(payload *tasks.SendEmailPayload) bool {
			return payload.To == req.Email
		})).Return(errors.New("email fail")).Once()

		err := uc.InviteMember(context.Background(), orgID, req, actorID)
		// Should log error but NOT fail request (soft failure for notification)
		assert.NoError(t, err)
	})
}

func TestRemoveMember_Extended(t *testing.T) {
	deps := &orgTestDeps{
		Repo:       new(orgMocks.MockOrganizationRepository),
		MemberRepo: new(orgMocks.MockOrganizationMemberRepository),
		TM:         new(mocking.MockWithTransactionManager),
		Enforcer:   new(permMocks.IEnforcer),
	}
	uc := setupMemberUseCase(deps)

	orgID := "org1"
	targetUserID := "owner1"
	actorID := "admin1"

	t.Run("Remove_Owner_Restriction", func(t *testing.T) {
		// Actor is admin, Target is owner
		deps.MemberRepo.On("FindMember", mock.Anything, orgID, actorID).Return(&entity.OrganizationMember{Role: "admin"}, nil).Once()
		deps.MemberRepo.On("FindMember", mock.Anything, orgID, targetUserID).Return(&entity.OrganizationMember{Role: "owner"}, nil).Once()

		err := uc.RemoveMember(context.Background(), orgID, targetUserID, actorID)
		assert.ErrorIs(t, err, exception.ErrForbidden)
	})
}
