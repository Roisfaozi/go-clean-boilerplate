package http

import (
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ProjectController struct {
	useCase  usecase.ProjectUseCase
	validate *validator.Validate
}

func NewProjectController(useCase usecase.ProjectUseCase, validate *validator.Validate) *ProjectController {
	return &ProjectController{
		useCase:  useCase,
		validate: validate,
	}
}

func (h *ProjectController) Create(c *gin.Context) {
	var req model.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
		return
	}

	userID := c.GetString("user_id")
	orgID := database.GetOrganizationID(c.Request.Context())

	res, err := h.useCase.CreateProject(c.Request.Context(), userID, orgID, req)
	if err != nil {
		response.HandleError(c, err, "failed to create project")
		return
	}
	response.Created(c, res)
}

func (h *ProjectController) GetAll(c *gin.Context) {
	orgID := database.GetOrganizationID(c.Request.Context())
	res, err := h.useCase.GetProjects(c.Request.Context(), orgID)
	if err != nil {
		response.HandleError(c, err, "failed to get projects")
		return
	}
	response.Success(c, res)
}

func (h *ProjectController) GetByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.useCase.GetProjectByID(c.Request.Context(), id)
	if err != nil {
		response.HandleError(c, err, "failed to get project")
		return
	}
	response.Success(c, res)
}

func (h *ProjectController) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
		return
	}

	res, err := h.useCase.UpdateProject(c.Request.Context(), id, req)
	if err != nil {
		response.HandleError(c, err, "failed to update project")
		return
	}
	response.Success(c, res)
}

func (h *ProjectController) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.useCase.DeleteProject(c.Request.Context(), id); err != nil {
		response.HandleError(c, err, "failed to delete project")
		return
	}
	response.Success(c, gin.H{"message": "project deleted successfully"})
}
