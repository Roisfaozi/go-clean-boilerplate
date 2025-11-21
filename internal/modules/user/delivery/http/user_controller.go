package http

import (
	"errors"
	"net/http"

	"github.com/Roisfaozi/casbin-db/internal/modules/user/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/Roisfaozi/casbin-db/internal/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	UserUseCase usecase.UserUseCase
	Log         *logrus.Logger
}

func NewUserHandler(userUseCase usecase.UserUseCase, log *logrus.Logger) *UserHandler {
	return &UserHandler{
		UserUseCase: userUseCase,
		Log:         log,
	}
}

// RegisterUser handles user registration
// @Summary      Register a new user
// @Description  Creates a new user account.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body model.RegisterUserRequest true "User Registration Details"
// @Success      201  {object}  response.WebResponse[model.UserResponse]
// @Failure      400  {object}  response.WebResponse[any] "Invalid request body"
// @Failure      409  {object}  response.WebResponse[any] "User with the same ID already exists"
// @Failure      500  {object}  response.WebResponse[any] "Internal server error"
// @Router       /users/register [post]
func (h *UserHandler) RegisterUser(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.RegisterUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.WithError(err).Error("failed to bind request body")
		response.BadRequest(c, errors.New("invalid request body"))
		return
	}

	user, err := h.UserUseCase.Create(ctx, &req)
	if err != nil {
		h.handleError(c, err, "failed to create user")
		return
	}

	response.Created(c, user)
}

// GetCurrentUser gets the currently authenticated user's information
// @Summary      Get current user
// @Description  Retrieves profile information for the currently authenticated user.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.WebResponse[model.UserResponse]
// @Failure      401  {object}  response.WebResponse[any] "Unauthorized"
// @Failure      404  {object}  response.WebResponse[any] "User not found"
// @Failure      500  {object}  response.WebResponse[any] "Internal server error"
// @Router       /users/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, errors.New("unauthorized"))
		return
	}

	req := &model.GetUserRequest{
		ID: userID.(string),
	}

	user, err := h.UserUseCase.Current(ctx, req)
	if err != nil {
		h.handleError(c, err, "failed to get current user")
		return
	}

	response.Success(c, user)
}

// UpdateUser updates user information
// @Summary      Update current user
// @Description  Updates the name or password for the currently authenticated user.
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.UpdateUserRequest true "Fields to update"
// @Success      200  {object}  response.WebResponse[model.UserResponse]
// @Failure      400  {object}  response.WebResponse[any] "Invalid request body"
// @Failure      401  {object}  response.WebResponse[any] "Unauthorized"
// @Failure      404  {object}  response.WebResponse[any] "User not found"
// @Failure      500  {object}  response.WebResponse[any] "Internal server error"
// @Router       /users/me [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, errors.New("unauthorized"))
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.WithError(err).Error("failed to bind request body")
		response.BadRequest(c, errors.New("invalid request body"))
		return
	}

	req.ID = userID.(string)

	user, err := h.UserUseCase.Update(ctx, &req)
	if err != nil {
		h.handleError(c, err, "failed to update user")
		return
	}

	response.Success(c, user)
}

// GetAllUsers retrieves all users
// @Summary      Get all users
// @Description  Retrieves a list of all users. Accessible only by admins/superadmins.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.WebResponse[[]model.UserResponse]
// @Failure      401  {object}  response.WebResponse[any] "Unauthorized"
// @Failure      403  {object}  response.WebResponse[any] "Forbidden"
// @Failure      500  {object}  response.WebResponse[any] "Internal server error"
// @Router       /users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	ctx := c.Request.Context()

	users, err := h.UserUseCase.GetAllUsers(ctx)
	if err != nil {
		h.handleError(c, err, "failed to get all users")
		return
	}

	response.Success(c, users)
}

// GetUserByID retrieves a single user by their ID
// @Summary      Get user by ID
// @Description  Retrieves profile information for a specific user by their ID. Accessible only by admins/superadmins.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  response.WebResponse[model.UserResponse]
// @Failure      401  {object}  response.WebResponse[any] "Unauthorized"
// @Failure      403  {object}  response.WebResponse[any] "Forbidden"
// @Failure      404  {object}  response.WebResponse[any] "User not found"
// @Failure      500  {object}  response.WebResponse[any] "Internal server error"
// @Router       /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	user, err := h.UserUseCase.GetUserByID(ctx, userID)
	if err != nil {
		h.handleError(c, err, "failed to get user by id")
		return
	}
	response.Success(c, user)
}

// DeleteUser deletes a user by their ID
// @Summary      Delete user by ID
// @Description  Deletes a specific user by their ID. Accessible only by admins/superadmins.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  response.WebResponse[any] "message: User deleted successfully"
// @Failure      401  {object}  response.WebResponse[any] "Unauthorized"
// @Failure      403  {object}  response.WebResponse[any] "Forbidden"
// @Failure      404  {object}  response.WebResponse[any] "User not found"
// @Failure      500  {object}  response.WebResponse[any] "Internal server error"
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	err := h.UserUseCase.DeleteUser(ctx, userID)
	if err != nil {
		h.handleError(c, err, "failed to delete user")
		return
	}

	response.Success(c, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) handleError(c *gin.Context, err error, message string) {
	h.Log.WithError(err).Error(message)

	switch {
	case errors.Is(err, exception.ErrBadRequest):
		response.BadRequest(c, err)
	case errors.Is(err, exception.ErrUnauthorized):
		response.Unauthorized(c, err)
	case errors.Is(err, exception.ErrForbidden):
		response.Forbidden(c, err)
	case errors.Is(err, exception.ErrNotFound):
		response.NotFound(c, err)
	case errors.Is(err, exception.ErrConflict):
		response.ErrorResponse(c, http.StatusConflict, err)
	default:
		response.InternalServerError(c, errors.New("internal server error"))
	}
}
