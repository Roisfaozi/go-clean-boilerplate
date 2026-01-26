package usecase

import (
	"io"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	permMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	storageMocks "github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/mocks"
	"github.com/sirupsen/logrus"
)

type userTestDeps struct {
	Repo     *mocks.MockUserRepository
	TM       *mocking.MockWithTransactionManager
	Enforcer *permMocks.IEnforcer
	AuditUC  *auditMocks.MockAuditUseCase
	AuthUC   *authMocks.MockAuthUseCase
	Storage  *storageMocks.MockProvider
}

func setupUserTest() (*userTestDeps, UserUseCase) {
	deps := &userTestDeps{
		Repo:     new(mocks.MockUserRepository),
		TM:       new(mocking.MockWithTransactionManager),
		Enforcer: new(permMocks.IEnforcer),
		AuditUC:  new(auditMocks.MockAuditUseCase),
		AuthUC:   new(authMocks.MockAuthUseCase),
		Storage:  new(storageMocks.MockProvider),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := NewUserUseCase(deps.TM, log, deps.Repo, deps.Enforcer, deps.AuditUC, deps.AuthUC, deps.Storage)

	return deps, uc
}
