package test

import (
	"io"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	mock_auth "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	mock_org "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	mock_permission "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	mock_user "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/sirupsen/logrus"
)

func TestRepro(t *testing.T) {
	jwtManager := jwt.NewJWTManager("secret", "refresh", 1, 1)
	log := logrus.New()
	log.SetOutput(io.Discard)

	_ = usecase.NewAuthUsecase(
		5,
		30*time.Minute,
		jwtManager,
		new(mock_auth.MockTokenRepository),
		new(mock_user.MockUserRepository),
		new(mock_org.MockOrganizationRepository),
		new(mocking.MockWithTransactionManager),
		log,
		new(mocking.MockManager),
		(*sse.Manager)(nil),
		new(mock_permission.IEnforcer),
		new(auditMocks.MockAuditUseCase),
		new(mocking.MockTaskDistributor),
		new(mock_auth.MockTicketManager),
	)
}
