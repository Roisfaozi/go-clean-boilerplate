package test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type projectTestDeps struct {
	Repo *mocks.MockProjectRepository
}

func setupProjectTest() (*projectTestDeps, usecase.ProjectUseCase) {
	deps := &projectTestDeps{
		Repo: new(mocks.MockProjectRepository),
	}

	uc := usecase.NewProjectUseCase(deps.Repo)

	return deps, uc
}

func TestProjectUseCase_Create_XSS_Sanitization(t *testing.T) {
	deps, uc := setupProjectTest()

	inputName := "<script>alert('XSS')</script>Project"
	inputDomain := "<img src=x onerror=alert(1)>example.com"

	// Expected sanitized output (using pkg.SanitizeString which uses html.EscapeString)
	expectedName := "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;Project"
	expectedDomain := "&lt;img src=x onerror=alert(1)&gt;example.com"

	req := model.CreateProjectRequest{
		Name:   inputName,
		Domain: inputDomain,
	}

	// We expect the repository to receive the SANITIZED values
	deps.Repo.On("Create", mock.Anything, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == expectedName && p.Domain == expectedDomain
	})).Return(nil)

	_, err := uc.CreateProject(context.Background(), "user1", "org1", req)

	// Since current implementation DOES NOT sanitize, this test is expected to fail.
	// However, if I want to "reproduce" the failure, I should assert no error,
	// but the mock expectation will fail because it receives unsanitized data.

	// Assert that the error is nil (meaning the function completed)
	// but the Mock assertion will fail at the end if the arguments didn't match.
	assert.NoError(t, err)

	// This will fail if the mock wasn't called with expected arguments
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Update_XSS_Sanitization(t *testing.T) {
	deps, uc := setupProjectTest()

	inputName := "<script>alert('update')</script>"
	expectedName := "&lt;script&gt;alert(&#39;update&#39;)&lt;/script&gt;"

	projectID := "proj1"
	existingProject := &entity.Project{
		ID: projectID,
		Name: "Original",
		Domain: "orig.com",
	}

	req := model.UpdateProjectRequest{
		Name: inputName,
	}

	deps.Repo.On("GetByID", mock.Anything, projectID).Return(existingProject, nil)

	// Expect sanitized name in Update
	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == expectedName
	})).Return(nil)

	_, err := uc.UpdateProject(context.Background(), projectID, req)

	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}
