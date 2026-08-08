package test

import (
	"io"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	mock_auth "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	mock_org "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	mock_user "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/sirupsen/logrus"
)

func TestRepro(t *testing.T) {
	jwtManager := jwt.NewJWTManager("secret", "refresh", 1, 1)
	log := logrus.New()
	log.SetOutput(io.Discard)

	_ = usecase.NewAuthUsecase(
		5,
		30*time.Minute,
		3,
		jwtManager,
		mock_auth.NewMockTokenRepository(t),
		mock_user.NewMockUserRepository(t),
		mock_org.NewMockOrganizationRepository(t),
		mocking.NewMockWithTransactionManager(t),
		log,
		mock_auth.NewMockNotificationPublisher(t),
		mock_auth.NewMockAuthzManager(t),
		mocking.NewMockTaskDistributor(t),
		mocking.NewMockTicketManager(t),
		nil,
	)
}
