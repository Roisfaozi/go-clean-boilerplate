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
	// Cast PresenceReader
	var pr usecase.PresenceReader = deps.PresenceReader

	return usecase.NewOrganizationMemberUseCase(log, deps.TM, deps.MemberRepo, deps.Repo, deps.InviteRepo, deps.UserRepo, deps.TaskDistributor, enf, pr, "http://frontend")
}

func TestCreateOrganization_Extended(t *testing.T) {
	deps := &orgTestDeps{
		Repo:       new(orgMocks.MockOrganizationRepository),
		MemberRepo: new(orgMocks.MockOrganizationMemberRepository),
		TM:         new(mocking.MockWithTransactionManager),
		Enforcer:   new(permMocks.IEnforcer),
	}
	uc := setupOrganizationUseCase(deps)

	req := &model.CreateOrganizationRequest{Name: "New Org", Slug: "new-org"}
	userID := "user1"

	t.Run("Transaction_Failure", func(t *testing.T) {
		deps.Repo.On("SlugExists", mock.Anything, "new-org").Return(false, nil).Once()
		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(exception.ErrInternalServer)

		deps.Repo.On("Create", mock.Anything, mock.Anything, "role:org-owner").Return(errors.New("db error")).Once()

		_, err := uc.CreateOrganization(context.Background(), userID, req)
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
		PresenceReader:  new(authMocks.MockPresenceReader),
	}
	uc := setupMemberUseCase(deps)

	orgID := "org1"
	req := &model.InviteMemberRequest{Email: "invitee@example.com", RoleID: "member"}
	actorID := "admin1"

	t.Run("Email_Send_Failure", func(t *testing.T) {
		_ = actorID // Unused in this test setup as controller usually handles actor authz logic or middleware

		// 1. Org Check
		deps.Repo.On("FindByID", mock.Anything, orgID).Return(&entity.Organization{ID: orgID, Name: "Test Org"}, nil).Once()

		// 2. User Check
		deps.UserRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, errors.New("user not found")).Once()
		deps.UserRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.Email == req.Email
		})).Return(nil).Once()

		// 3. Membership Check
		deps.MemberRepo.On("CheckMembership", mock.Anything, orgID, mock.Anything).Return(false, nil).Once()
		deps.MemberRepo.On("GetMemberStatus", mock.Anything, orgID, mock.Anything).Return("", nil).Once()

		// 4. Add Member
		deps.MemberRepo.On("AddMember", mock.Anything, mock.Anything).Return(nil).Once()

		// 5. Invitation Token
		deps.InviteRepo.On("DeleteByEmailAndOrg", mock.Anything, req.Email, orgID).Return(nil).Once()
		deps.InviteRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		// Transaction
		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(nil)

		// Email failure
		deps.TaskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.MatchedBy(func(payload *tasks.SendEmailPayload) bool {
			return payload.To == req.Email
		})).Return(errors.New("email fail")).Once()

		_, err := uc.InviteMember(context.Background(), orgID, req)
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
		InviteRepo: new(orgMocks.MockInvitationRepository), // Added missing repo
		UserRepo:   new(userMocks.MockUserRepository),
		TaskDistributor: new(mocking.MockTaskDistributor),
		PresenceReader: new(authMocks.MockPresenceReader),
	}
	uc := setupMemberUseCase(deps)

	orgID := "org1"
	targetUserID := "owner1"

	t.Run("Remove_Owner_Restriction", func(t *testing.T) {
		// Mock checks inside transaction
		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.ErrorIs(t, err, exception.ErrForbidden)
			}).Return(exception.ErrForbidden)

		// 1. Check if member exists
		deps.MemberRepo.On("CheckMembership", mock.Anything, orgID, targetUserID).Return(true, nil).Once()

		// 2. Check if owner
		deps.Repo.On("FindByID", mock.Anything, orgID).Return(&entity.Organization{ID: orgID, OwnerID: targetUserID}, nil).Once()

		err := uc.RemoveMember(context.Background(), orgID, targetUserID)
		assert.ErrorIs(t, err, exception.ErrForbidden)
	})
}
