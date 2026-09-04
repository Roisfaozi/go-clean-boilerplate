package http

import (
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
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

// Create godoc
// @Summary      Create project
// @Description  Creates a new project within the active organization context.
// @Tags         projects
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID header string true "Organization ID"
// @Param        request body model.CreateProjectRequest true "Create Project Request"
// @Success      201  {object}  response.SwaggerSuccessResponseWrapper{data=model.ProjectResponse} "Project created successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body or missing organization ID"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /projects [post]
func (h *ProjectController) Create(c *gin.Context) {
	var req model.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, "validation error")
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

func (h *ProjectController) HTTPCreate(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProjectRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, "validation error")
		return
	}

	userID, _ := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	orgID := database.GetOrganizationID(r.Context())

	res, err := h.useCase.CreateProject(r.Context(), userID, orgID, req)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to create project")
		return
	}
	response.WriteSuccess(w, http.StatusCreated, res)
}

// GetAll godoc
// @Summary      Get all projects
// @Description  Returns a list of all projects belonging to the active organization.
// @Tags         projects
// @Security     BearerAuth
// @Produce      json
// @Param        X-Organization-ID header string true "Organization ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=[]model.ProjectResponse} "Projects retrieved successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Missing organization ID"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /projects [get]
func (h *ProjectController) GetAll(c *gin.Context) {
	orgID := database.GetOrganizationID(c.Request.Context())
	res, err := h.useCase.GetProjects(c.Request.Context(), orgID)
	if err != nil {
		response.HandleError(c, err, "failed to get projects")
		return
	}
	response.Success(c, res)
}

func (h *ProjectController) HTTPGetAll(w http.ResponseWriter, r *http.Request) {
	orgID := database.GetOrganizationID(r.Context())
	res, err := h.useCase.GetProjects(r.Context(), orgID)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to get projects")
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}

// GetByID godoc
// @Summary      Get project by ID
// @Description  Returns detailed information about a specific project.
// @Tags         projects
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Project ID"
// @Param        X-Organization-ID header string true "Organization ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=model.ProjectResponse} "Project retrieved successfully"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "Project not found"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /projects/{id} [get]
func (h *ProjectController) GetByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.useCase.GetProjectByID(c.Request.Context(), id)
	if err != nil {
		response.HandleError(c, err, "failed to get project")
		return
	}
	response.Success(c, res)
}

func (h *ProjectController) HTTPGetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	res, err := h.useCase.GetProjectByID(r.Context(), id)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to get project")
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}

// Update godoc
// @Summary      Update project
// @Description  Updates an existing project's details.
// @Tags         projects
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Project ID"
// @Param        X-Organization-ID header string true "Organization ID"
// @Param        request body model.UpdateProjectRequest true "Update Project Request"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=model.ProjectResponse} "Project updated successfully"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "Project not found"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /projects/{id} [put]
func (h *ProjectController) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, "validation error")
		return
	}

	res, err := h.useCase.UpdateProject(c.Request.Context(), id, req)
	if err != nil {
		response.HandleError(c, err, "failed to update project")
		return
	}
	response.Success(c, res)
}

func (h *ProjectController) HTTPUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req model.UpdateProjectRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, "validation error")
		return
	}

	res, err := h.useCase.UpdateProject(r.Context(), id, req)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to update project")
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}

// Delete godoc
// @Summary      Delete project
// @Description  Soft deletes a project from the active organization.
// @Tags         projects
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Project ID"
// @Param        X-Organization-ID header string true "Organization ID"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "Project deleted successfully"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "Project not found"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /projects/{id} [delete]
func (h *ProjectController) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.useCase.DeleteProject(c.Request.Context(), id); err != nil {
		response.HandleError(c, err, "failed to delete project")
		return
	}
	response.Success(c, gin.H{"message": "project deleted successfully"})
}

func (h *ProjectController) HTTPDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.useCase.DeleteProject(r.Context(), id); err != nil {
		response.WriteHTTPError(w, err, "failed to delete project")
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "project deleted successfully"})
}
