package http

import (
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type UserController struct {
	UserUseCase usecase.UserUseCase
	Log         *logrus.Logger
	validate    *validator.Validate
}

func NewUserController(userUseCase usecase.UserUseCase, log *logrus.Logger, validate *validator.Validate) *UserController {
	return &UserController{
		UserUseCase: userUseCase,
		Log:         log,
		validate:    validate,
	}
}

// RegisterUser handles user registration
// @Summary      Register a new user
// @Description  Creates a new user account.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body model.RegisterUserRequest true "User Registration Details"
// @Success      201  {object}  response.SwaggerUserResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      409  {object}  response.SwaggerErrorResponseWrapper "User with the same ID already exists"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /users/register [post]
func (h *UserController) RegisterUser(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.RegisterUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.WithError(err).Error("failed to bind request body")
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		h.Log.WithError(err).Error(msg)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	req.IPAddress = c.ClientIP()
	req.UserAgent = c.Request.UserAgent()

	user, err := h.UserUseCase.Create(ctx, &req)
	if err != nil {
		h.Log.WithError(err).Error("failed to create user")
		response.HandleError(c, err, "failed to create user")
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
// @Success      200  {object}  response.SwaggerUserResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "User not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /users/me [get]
func (h *UserController) GetCurrentUser(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, exception.ErrUnauthorized, "unauthorized")
		return
	}

	req := &model.GetUserRequest{
		ID: userID.(string),
	}

	user, err := h.UserUseCase.Current(ctx, req)
	if err != nil {
		h.Log.WithError(err).Error("failed to get current user")
		response.HandleError(c, err, "failed to get current user")
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
// @Success      200  {object}  response.SwaggerUserResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "User not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /users/me [put]
func (h *UserController) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, exception.ErrUnauthorized, "unauthorized")
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.WithError(err).Error("failed to bind request body")
		response.BadRequest(c, err, "invalid request body")
		return
	}

	req.ID = userID.(string)

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	req.IPAddress = c.ClientIP()
	req.UserAgent = c.Request.UserAgent()

	user, err := h.UserUseCase.Update(ctx, &req)
	if err != nil {
		h.Log.WithError(err).Error("failed to update user")
		response.HandleError(c, err, "failed to update user")
		return
	}

	response.Success(c, user)
}

// UpdateAvatar handles user avatar upload
// @Summary      Upload avatar
// @Description  Uploads a new avatar image for the current user.
// @Tags         users
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        avatar formData file true "Avatar image file"
// @Success      200  {object}  response.SwaggerUserResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid file"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /users/me/avatar [patch]
func (h *UserController) UpdateAvatar(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, exception.ErrUnauthorized, "unauthorized")
		return
	}

	// 1. Get file from form
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		response.BadRequest(c, err, "failed to get avatar file from request")
		return
	}
	defer func() {
		_ = file.Close()
	}()

	// 2. Validate file size (max 2MB)
	if header.Size > 2*1024*1024 {
		response.BadRequest(c, errors.New("file too large"), "avatar size must be less than 2MB")
		return
	}

	// 3. Call UseCase
	user, err := h.UserUseCase.UpdateAvatar(ctx, userID.(string), file, header.Filename, header.Header.Get("Content-Type"))
	if err != nil {
		h.Log.WithError(err).Error("failed to update avatar")
		response.HandleError(c, err, "failed to update avatar")
		return
	}

	response.Success(c, user)
}

// UpdateUserStatus updates user status (active, suspended, banned)
// @Summary      Update user status
// @Description  Updates the status of a specific user. Accessible only by admins/superadmins.
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id     path      string  true  "User ID"
// @Param        request body model.UpdateUserStatusRequest true "New Status"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      403  {object}  response.SwaggerErrorResponseWrapper "Forbidden"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "User not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /users/{id}/status [patch]
func (h *UserController) UpdateUserStatus(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	var req model.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.WithError(err).Error("failed to bind request body")
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	err := h.UserUseCase.UpdateStatus(ctx, userID, req.Status)
	if err != nil {
		h.Log.WithError(err).Error("failed to update user status")
		response.HandleError(c, err, "failed to update user status")
		return
	}

	response.Success(c, gin.H{"message": "User status updated successfully"})
}

// GetAllUsers retrieves all users with pagination and filtering
// @Summary      Get all users
// @Description  Retrieves a list of all users with optional pagination and filtering. Accessible only by admins/superadmins.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        page      query     int     false  "Page number (default 1)"
// @Param        limit     query     int     false  "Items per page (default 10)"
// @Param        username  query     string  false  "Filter by username"
// @Param        email     query     string  false  "Filter by email"
// @Success      200  {object}  response.SwaggerUserListResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid query parameters"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      403  {object}  response.SwaggerErrorResponseWrapper "Forbidden"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /users [get]
func (h *UserController) GetAllUsers(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.GetUserListRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		h.Log.WithError(err).Error("failed to bind query parameters")
		response.BadRequest(c, err, "invalid query parameters")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		h.Log.WithError(err).Error(msg)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	users, total, err := h.UserUseCase.GetAllUsers(ctx, &req)
	if err != nil {
		h.Log.WithError(err).Error("failed to get all users")
		response.HandleError(c, err, "failed to get all users")
		return
	}

	response.SuccessResponseWithPaging(c, users, &response.PageMetadata{
		Page:  req.Page,
		Limit: req.Limit,
		Total: total,
	})
}

// GetUserByID retrieves a single user by their ID
// @Summary      Get user by ID
// @Description  Retrieves profile information for a specific user by their ID. Accessible only by admins/superadmins.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  response.SwaggerUserResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      403  {object}  response.SwaggerErrorResponseWrapper "Forbidden"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "User not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /users/{id} [get]
func (h *UserController) GetUserByID(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	user, err := h.UserUseCase.GetUserByID(ctx, userID)
	if err != nil {
		h.Log.WithError(err).Error("failed to get user by id")
		response.HandleError(c, err, "failed to get user by id")
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
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      403  {object}  response.SwaggerErrorResponseWrapper "Forbidden"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper "User not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /users/{id} [delete]
func (h *UserController) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()
	// Get actor's UserID from context
	actorUserID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, exception.ErrUnauthorized, "Please login to perform this action")
		return
	}

	var req model.DeleteUserRequest
	req.ID = c.Param("id")
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.Request.UserAgent()

	err := h.UserUseCase.DeleteUser(ctx, actorUserID.(string), &req)
	if err != nil {
		h.Log.WithError(err).Error("failed to delete user")
		response.HandleError(c, err, "failed to delete user")
		return
	}

	response.Success(c, gin.H{"message": "User deleted successfully"})
}

// GetUsersDynamic retrieves users based on dynamic filters and sorting via POST request body
// @Summary      Get users with dynamic filters
// @Description  Retrieves a list of users based on dynamic filter and sort criteria provided in the request body.
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        filter body querybuilder.DynamicFilter true "Dynamic filter and sort criteria"
// @Success      200  {object}  response.SwaggerUserListResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body or filter criteria"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      403  {object}  response.SwaggerErrorResponseWrapper "Forbidden"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /users/search [post]
func (h *UserController) GetUsersDynamic(c *gin.Context) {
	ctx := c.Request.Context()
	var filter querybuilder.DynamicFilter

	if err := c.ShouldBindJSON(&filter); err != nil {
		h.Log.WithError(err).Error("failed to bind dynamic filter request body")
		response.BadRequest(c, err, "invalid request body for dynamic filter")
		return
	}

	if err := h.validate.Struct(filter); err != nil {
		msg := validation.FormatValidationErrors(err)
		h.Log.WithError(err).Error(msg)
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	users, total, err := h.UserUseCase.GetAllUsersDynamic(ctx, &filter)
	if err != nil {
		h.Log.WithError(err).Error("failed to get users dynamically")
		response.HandleError(c, err, "failed to retrieve users")
		return
	}

	response.SuccessResponseWithPaging(c, users, &response.PageMetadata{
		Page:  filter.Page,
		Limit: filter.PageSize,
		Total: total,
	})
}
