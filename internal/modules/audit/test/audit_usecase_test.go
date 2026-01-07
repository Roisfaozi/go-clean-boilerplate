package test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type auditTestDeps struct {
	Repo *auditMocks.MockAuditRepository
}

func setupAuditTest() (*auditTestDeps, usecase.AuditUseCase) {
	deps := &auditTestDeps{
		Repo: new(auditMocks.MockAuditRepository),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewAuditUseCase(deps.Repo, log)
	return deps, uc
}

func TestAuditUseCase_LogActivity_Success(t *testing.T) {
	deps, uc := setupAuditTest()
	req := model.CreateAuditLogRequest{
		UserID: "u1",
		Action: "CREATE",
		Entity: "User",
	}

	deps.Repo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AuditLog) bool {
		return l.UserID == "u1" && l.Action == "CREATE"
	})).Return(nil)

	err := uc.LogActivity(context.Background(), req)

	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}

func TestAuditUseCase_LogActivity_ValidationError(t *testing.T) {
	_, uc := setupAuditTest()
	// Missing UserID
	req := model.CreateAuditLogRequest{
		Action: "CREATE",
		Entity: "User",
	}

	err := uc.LogActivity(context.Background(), req)

	assert.ErrorContains(t, err, "missing required fields")
}

func TestAuditUseCase_LogActivity_RepoError(t *testing.T) {
	deps, uc := setupAuditTest()
	req := model.CreateAuditLogRequest{UserID: "u1", Action: "A", Entity: "E"}

	deps.Repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	err := uc.LogActivity(context.Background(), req)

	assert.Error(t, err)
}

func TestAuditUseCase_GetLogsDynamic_Success(t *testing.T) {
	deps, uc := setupAuditTest()
	logs := []*entity.AuditLog{
		{UserID: "u1", Action: "A", OldValues: "{}", NewValues: "{}"},
	}
	filter := &querybuilder.DynamicFilter{}

	deps.Repo.On("FindAllDynamic", mock.Anything, filter).Return(logs, nil)

	res, err := uc.GetLogsDynamic(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestAuditUseCase_GetLogsDynamic_RepoError(t *testing.T) {
	deps, uc := setupAuditTest()
	filter := &querybuilder.DynamicFilter{}

	deps.Repo.On("FindAllDynamic", mock.Anything, filter).Return(nil, errors.New("db error"))

	res, err := uc.GetLogsDynamic(context.Background(), filter)

	assert.Error(t, err)
	assert.Nil(t, res)
}
