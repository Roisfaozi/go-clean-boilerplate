package http

import (
	"net/http"
	"strconv"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
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

// HTTPRegisterUser handles user registration
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
func (h *UserController) HTTPRegisterUser(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterUserRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		h.Log.WithError(err).Error("failed to bind request body")
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	req.IPAddress = r.RemoteAddr
	req.UserAgent = r.UserAgent()
	user, err := h.UserUseCase.Create(r.Context(), &req)
	if err != nil {
		h.Log.WithError(err).Error("failed to create user")
		response.WriteHTTPError(w, err, "failed to create user")
		return
	}
	response.WriteSuccess(w, http.StatusCreated, user)
}

// HTTPGetCurrentUser gets the currently authenticated user's information
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
func (h *UserController) HTTPGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || userID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "unauthorized")
		return
	}
	user, err := h.UserUseCase.Current(r.Context(), &model.GetUserRequest{ID: userID})
	if err != nil {
		h.Log.WithError(err).Error("failed to get current user")
		response.WriteHTTPError(w, err, "failed to get current user")
		return
	}
	response.WriteSuccess(w, http.StatusOK, user)
}

// HTTPUpdateUser updates user information
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
func (h *UserController) HTTPUpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || userID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "unauthorized")
		return
	}
	var req model.UpdateUserRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		h.Log.WithError(err).Error("failed to bind request body")
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	req.ID = userID
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	req.IPAddress = r.RemoteAddr
	req.UserAgent = r.UserAgent()
	user, err := h.UserUseCase.Update(r.Context(), &req)
	if err != nil {
		h.Log.WithError(err).Error("failed to update user")
		response.WriteHTTPError(w, err, "failed to update user")
		return
	}
	response.WriteSuccess(w, http.StatusOK, user)
}

// HTTPUpdateAvatar handles user avatar upload
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
func (h *UserController) HTTPUpdateAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || userID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "unauthorized")
		return
	}
	if err := r.ParseMultipartForm(2 * 1024 * 1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "avatar size must be less than 2MB")
		return
	}
	file, header, err := r.FormFile("avatar")
	if err != nil {
		h.Log.WithError(err).Error("failed to get avatar file from request")
		response.WriteHTTPError(w, exception.ErrBadRequest, "failed to get avatar file from request")
		return
	}
	defer func() {
		_ = file.Close()
	}()
	if header.Size > 2*1024*1024 {
		response.WriteHTTPError(w, exception.ErrBadRequest, "avatar size must be less than 2MB")
		return
	}
	user, err := h.UserUseCase.UpdateAvatar(r.Context(), userID, file, header.Filename, header.Header.Get("Content-Type"))
	if err != nil {
		h.Log.WithError(err).Error("failed to update avatar")
		response.WriteHTTPError(w, err, "failed to update avatar")
		return
	}
	response.WriteSuccess(w, http.StatusOK, user)
}

// HTTPUpdateUserStatus updates user status (active, suspended, banned)
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
func (h *UserController) HTTPUpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	var req model.UpdateUserStatusRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		h.Log.WithError(err).Error("failed to bind request body")
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	err := h.UserUseCase.UpdateStatus(r.Context(), userID, req.Status)
	if err != nil {
		h.Log.WithError(err).Error("failed to update user status")
		response.WriteHTTPError(w, err, "failed to update user status")
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "User status updated successfully"})
}

// HTTPGetAllUsers retrieves all users with pagination and filtering
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
func (h *UserController) HTTPGetAllUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var page, limit int
	if pageStr := query.Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			h.Log.WithError(err).Error("failed to bind query parameters")
			response.WriteHTTPError(w, exception.ErrBadRequest, "invalid query parameters")
			return
		}
		page = p
	}
	if limitStr := query.Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			h.Log.WithError(err).Error("failed to bind query parameters")
			response.WriteHTTPError(w, exception.ErrBadRequest, "invalid query parameters")
			return
		}
		limit = l
	}
	req := model.GetUserListRequest{
		Page:     page,
		Limit:    limit,
		Username: query.Get("username"),
		Email:    query.Get("email"),
	}
	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	users, total, err := h.UserUseCase.GetAllUsers(r.Context(), &req)
	if err != nil {
		h.Log.WithError(err).Error("failed to get all users")
		response.WriteHTTPError(w, err, "failed to get all users")
		return
	}
	response.WriteSuccessWithPaging(w, users, &response.PageMetadata{
		Page:  req.Page,
		Limit: req.Limit,
		Total: total,
	})
}

// HTTPGetUserByID retrieves a single user by their ID
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
func (h *UserController) HTTPGetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	user, err := h.UserUseCase.GetUserByID(r.Context(), userID)
	if err != nil {
		h.Log.WithError(err).Error("failed to get user by id")
		response.WriteHTTPError(w, err, "failed to get user by id")
		return
	}
	response.WriteSuccess(w, http.StatusOK, user)
}

// HTTPDeleteUser deletes a user by their ID
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
func (h *UserController) HTTPDeleteUser(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || actorUserID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "Please login to perform this action")
		return
	}
	req := model.DeleteUserRequest{
		ID:        r.PathValue("id"),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}
	err := h.UserUseCase.DeleteUser(r.Context(), actorUserID, &req)
	if err != nil {
		h.Log.WithError(err).Error("failed to delete user")
		response.WriteHTTPError(w, err, "failed to delete user")
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// HTTPGetUsersDynamic retrieves users based on dynamic filters and sorting via POST request body
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
func (h *UserController) HTTPGetUsersDynamic(w http.ResponseWriter, r *http.Request) {
	var filter querybuilder.DynamicFilter
	if err := response.DecodeJSON(r, &filter, 1024*1024); err != nil {
		h.Log.WithError(err).Error("failed to bind dynamic filter request body")
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body for dynamic filter")
		return
	}
	if err := h.validate.Struct(filter); err != nil {
		msg := validation.FormatValidationErrors(err)
		h.Log.WithError(err).Error(msg)
		response.WriteHTTPError(w, exception.ErrValidationError, msg)
		return
	}
	users, total, err := h.UserUseCase.GetAllUsersDynamic(r.Context(), &filter)
	if err != nil {
		h.Log.WithError(err).Error("failed to get users dynamically")
		response.WriteHTTPError(w, err, "failed to retrieve users")
		return
	}
	response.WriteSuccessWithPaging(w, users, &response.PageMetadata{
		Page:  filter.Page,
		Limit: filter.PageSize,
		Total: total,
	})
}
