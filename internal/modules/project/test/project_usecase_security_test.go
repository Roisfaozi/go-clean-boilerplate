package test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper function duplicated here because it is not exported in the other test file
// unless I export it there. But to avoid modifying existing test files unnecessarily,
// I will create a local setup helper.
func setupProjectSecurityTest() (*mocks.MockProjectRepository, usecase.ProjectUseCase) {
	repo := new(mocks.MockProjectRepository)
	uc := usecase.NewProjectUseCase(repo)
	return repo, uc
}

func TestProjectUseCase_Create_XSS_Sanitization(t *testing.T) {
	// 1. Setup
	repo, uc := setupProjectSecurityTest()
	ctx := context.Background()

	// 2. Define payload with XSS vector
	xssName := "<script>alert('xss')</script>Project"
	xssDomain := "<img src=x onerror=alert(1)>domain.com"

	// Expected sanitized values (using html.EscapeString behavior via pkg.SanitizeString)
	safeName := pkg.SanitizeString(xssName)
	safeDomain := pkg.SanitizeString(xssDomain)

	// 3. Mock Expectation: Repository should receive SANITIZED values
	// We use MatchedBy to verify the arguments passed to the repository
	repo.On("Create", ctx, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == safeName && p.Domain == safeDomain
	})).Return(nil).Once()

	req := model.CreateProjectRequest{Name: xssName, Domain: xssDomain}

	// 4. Execute
	result, err := uc.CreateProject(ctx, "user-1", "org-1", req)

	// 5. Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, safeName, result.Name)
	assert.Equal(t, safeDomain, result.Domain)
	repo.AssertExpectations(t)
}

func TestProjectUseCase_Update_XSS_Sanitization(t *testing.T) {
	// 1. Setup
	repo, uc := setupProjectSecurityTest()
	ctx := context.Background()

	existing := &entity.Project{
		ID: "p1", OrganizationID: "org-1", UserID: "u1",
		Name: "Old Name", Domain: "old.com", Status: "active",
	}

	// 2. Define payload
	xssName := "<b>Bold</b>"
	safeName := pkg.SanitizeString(xssName)

	repo.On("GetByID", ctx, "p1").Return(existing, nil).Once()

	// 3. Mock Expectation: Repository should receive SANITIZED Name
	repo.On("Update", ctx, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == safeName
	})).Return(nil).Once()

	req := model.UpdateProjectRequest{Name: xssName}

	// 4. Execute
	result, err := uc.UpdateProject(ctx, "p1", req)

	// 5. Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, safeName, result.Name)
	repo.AssertExpectations(t)
}
