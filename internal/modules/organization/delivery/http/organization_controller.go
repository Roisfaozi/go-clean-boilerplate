package http

import (
	"errors"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type OrganizationController struct {
	OrgUseCase    usecase.OrganizationUseCase
	MemberUseCase usecase.OrganizationMemberUseCase
	Log           *logrus.Logger
	validate      *validator.Validate
}

func NewOrganizationController(
	orgUseCase usecase.OrganizationUseCase,
	memberUseCase usecase.OrganizationMemberUseCase,
	log *logrus.Logger,
	validate *validator.Validate,
) *OrganizationController {
	return &OrganizationController{
		OrgUseCase:    orgUseCase,
		MemberUseCase: memberUseCase,
		Log:           log,
		validate:      validate,
	}
}

// HTTPCreateOrganization creates a new organization
// @Summary      Create organization
// @Description  Creates a new organization with the current user as owner
// @Tags         organizations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      model.CreateOrganizationRequest  true  "Organization creation request"
// @Success      201      {object}  response.SwaggerSuccessResponseWrapper{data=model.OrganizationResponse}
// @Failure      400      {object}  response.SwaggerErrorResponseWrapper  "Validation error"
// @Failure      401      {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      409      {object}  response.SwaggerErrorResponseWrapper  "Slug already exists"
// @Failure      500      {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations [post]
func (ctrl *OrganizationController) HTTPCreateOrganization(w http.ResponseWriter, r *http.Request) {
	userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || userID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	var request model.CreateOrganizationRequest
	if err := response.DecodeJSON(r, &request, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	request.Sanitize()
	if err := ctrl.validate.Struct(&request); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	result, err := ctrl.OrgUseCase.CreateOrganization(r.Context(), userID, &request)
	if err != nil {
		if errors.Is(err, exception.ErrConflict) {
			response.WriteError(w, http.StatusConflict, err, "organization slug already exists")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to create organization")
		response.WriteHTTPError(w, err, "failed to create organization")
		return
	}
	response.WriteSuccess(w, http.StatusCreated, result)
}

// HTTPGetOrganization retrieves an organization by ID
// @Summary      Get organization by ID
// @Description  Retrieves organization details by its ID
// @Tags         organizations
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=model.OrganizationResponse}
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper  "Organization not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/{id} [get]
func (ctrl *OrganizationController) HTTPGetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	userID, _ := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	ctx := usecase.WithActorUserID(r.Context(), userID)
	result, err := ctrl.OrgUseCase.GetOrganization(ctx, orgID)
	if err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			response.WriteHTTPError(w, exception.ErrNotFound, "organization not found")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to get organization")
		response.WriteHTTPError(w, err, "failed to get organization")
		return
	}
	response.WriteSuccess(w, http.StatusOK, result)
}

// HTTPGetOrganizationBySlug retrieves an organization by slug
// @Summary      Get organization by slug
// @Description  Retrieves organization details by its slug
// @Tags         organizations
// @Security     BearerAuth
// @Produce      json
// @Param        slug  path      string  true  "Organization Slug"
// @Success      200   {object}  response.SwaggerSuccessResponseWrapper{data=model.OrganizationResponse}
// @Failure      401   {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      404   {object}  response.SwaggerErrorResponseWrapper  "Organization not found"
// @Failure      500   {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/slug/{slug} [get]
func (ctrl *OrganizationController) HTTPGetOrganizationBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	userID, _ := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	ctx := usecase.WithActorUserID(r.Context(), userID)
	result, err := ctrl.OrgUseCase.GetOrganizationBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			response.WriteHTTPError(w, exception.ErrNotFound, "organization not found")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to get organization by slug")
		response.WriteHTTPError(w, err, "failed to get organization")
		return
	}
	response.WriteSuccess(w, http.StatusOK, result)
}

// HTTPUpdateOrganization updates organization details
// @Summary      Update organization
// @Description  Updates organization details (name, settings)
// @Tags         organizations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                            true  "Organization ID"
// @Param        request  body      model.UpdateOrganizationRequest   true  "Update request"
// @Success      200      {object}  response.SwaggerSuccessResponseWrapper{data=model.OrganizationResponse}
// @Failure      400      {object}  response.SwaggerErrorResponseWrapper  "Validation error"
// @Failure      401      {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      404      {object}  response.SwaggerErrorResponseWrapper  "Organization not found"
// @Failure      500      {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/{id} [put]
func (ctrl *OrganizationController) HTTPUpdateOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	userID, _ := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	var request model.UpdateOrganizationRequest
	if err := response.DecodeJSON(r, &request, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	request.Sanitize()
	if err := ctrl.validate.Struct(&request); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	ctx := usecase.WithActorUserID(r.Context(), userID)
	result, err := ctrl.OrgUseCase.UpdateOrganization(ctx, orgID, &request)
	if err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			response.WriteHTTPError(w, exception.ErrNotFound, "organization not found")
			return
		}
		if errors.Is(err, exception.ErrForbidden) {
			response.WriteHTTPError(w, exception.ErrForbidden, "you do not have permission to manage this organization")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to update organization")
		response.WriteHTTPError(w, err, "failed to update organization")
		return
	}
	response.WriteSuccess(w, http.StatusOK, result)
}

// HTTPDeleteOrganization deletes an organization
// @Summary      Delete organization
// @Description  Deletes an organization (owner only)
// @Tags         organizations
// @Security     BearerAuth
// @Produce      json
// @Param        id  path      string  true  "Organization ID"
// @Success      200 {object}  response.SwaggerSuccessResponseWrapper
// @Failure      401 {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      403 {object}  response.SwaggerErrorResponseWrapper  "Forbidden (not owner)"
// @Failure      404 {object}  response.SwaggerErrorResponseWrapper  "Organization not found"
// @Failure      500 {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/{id} [delete]
func (ctrl *OrganizationController) HTTPDeleteOrganization(w http.ResponseWriter, r *http.Request) {
	userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || userID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	orgID := r.PathValue("id")
	err := ctrl.OrgUseCase.DeleteOrganization(r.Context(), orgID, userID)
	if err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			response.WriteHTTPError(w, exception.ErrNotFound, "organization not found")
			return
		}
		if errors.Is(err, exception.ErrForbidden) {
			response.WriteHTTPError(w, exception.ErrForbidden, "only the owner can delete this organization")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to delete organization")
		response.WriteHTTPError(w, err, "failed to delete organization")
		return
	}
	response.WriteSuccess(w, http.StatusOK, nil)
}

// HTTPRestoreOrganization restores a soft-deleted organization.
func (ctrl *OrganizationController) HTTPRestoreOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || userID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	role, _ := delivery.GetContextString(r.Context(), delivery.RoleKey)
	ctx := usecase.WithActorUserID(r.Context(), userID)
	ctx = usecase.WithActorRole(ctx, role)
	result, err := ctrl.OrgUseCase.RestoreOrganization(ctx, orgID)
	if err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			response.WriteHTTPError(w, exception.ErrNotFound, "organization not found")
			return
		}
		if errors.Is(err, exception.ErrForbidden) {
			response.WriteHTTPError(w, exception.ErrForbidden, "only superadmin can restore this organization")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to restore organization")
		response.WriteHTTPError(w, err, "failed to restore organization")
		return
	}
	response.WriteSuccess(w, http.StatusOK, result)
}

// HTTPHardDeleteOrganization permanently deletes a previously soft-deleted organization.
func (ctrl *OrganizationController) HTTPHardDeleteOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || userID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	role, _ := delivery.GetContextString(r.Context(), delivery.RoleKey)
	ctx := usecase.WithActorUserID(r.Context(), userID)
	ctx = usecase.WithActorRole(ctx, role)
	err := ctrl.OrgUseCase.HardDeleteOrganization(ctx, orgID)
	if err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			response.WriteHTTPError(w, exception.ErrNotFound, "organization not found")
			return
		}
		if errors.Is(err, exception.ErrForbidden) {
			response.WriteHTTPError(w, exception.ErrForbidden, "only superadmin can hard delete this organization")
			return
		}
		if errors.Is(err, exception.ErrBadRequest) {
			response.WriteHTTPError(w, exception.ErrBadRequest, "organization must be soft-deleted before hard delete")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to hard delete organization")
		response.WriteHTTPError(w, err, "failed to hard delete organization")
		return
	}
	response.WriteSuccess(w, http.StatusOK, nil)
}

// HTTPGetMyOrganizations retrieves organizations for the current user
// @Summary      Get my organizations
// @Description  Retrieves all organizations the current user is a member of
// @Tags         organizations
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=model.UserOrganizationsResponse}
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/me [get]
func (ctrl *OrganizationController) HTTPGetMyOrganizations(w http.ResponseWriter, r *http.Request) {
	userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || userID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	result, err := ctrl.OrgUseCase.GetUserOrganizations(r.Context(), userID)
	if err != nil {
		ctrl.Log.WithError(err).Error("Failed to get user organizations")
		response.WriteHTTPError(w, err, "failed to get organizations")
		return
	}
	response.WriteSuccess(w, http.StatusOK, result)
}

// HTTPAcceptInvitation accounts an invitation
// @Summary      Accept invitation
// @Description  Accepts an invitation to join an organization. If user is new (shadow), sets password and activates account.
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        request body model.AcceptInvitationRequest true "Invitation Acceptance Request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper   "Success"
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper  "Bad Request"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/invitations/accept [post]
func (ctrl *OrganizationController) HTTPAcceptInvitation(w http.ResponseWriter, r *http.Request) {
	var request model.AcceptInvitationRequest
	if err := response.DecodeJSON(r, &request, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "Invalid request body")
		return
	}
	if request.Token == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "Validation failed")
		return
	}
	if err := ctrl.validate.Struct(&request); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "Validation failed")
		return
	}
	if err := ctrl.MemberUseCase.AcceptInvitation(r.Context(), &request); err != nil {
		response.WriteHTTPError(w, err, "Failed to accept invitation")
		return
	}
	response.WriteSuccess(w, http.StatusOK, nil)
}

// HTTPInviteMember invites a user to join an organization
// @Summary      Invite member to organization
// @Description  Invites a user by user ID to join an organization
// @Tags         organizations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                      true  "Organization ID"
// @Param        request  body      model.InviteMemberRequest   true  "Invite member request"
// @Success      201      {object}  response.SwaggerSuccessResponseWrapper{data=model.MemberResponse}
// @Failure      400      {object}  response.SwaggerErrorResponseWrapper  "Validation error"
// @Failure      401      {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      403      {object}  response.SwaggerErrorResponseWrapper  "Forbidden"
// @Failure      404      {object}  response.SwaggerErrorResponseWrapper  "Organization not found"
// @Failure      409      {object}  response.SwaggerErrorResponseWrapper  "User already a member"
// @Failure      500      {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/{id}/members/invite [post]
func (ctrl *OrganizationController) HTTPInviteMember(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || userID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	var request model.InviteMemberRequest
	if err := response.DecodeJSON(r, &request, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := ctrl.validate.Struct(&request); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	ctx := usecase.WithActorUserID(r.Context(), userID)
	result, err := ctrl.MemberUseCase.InviteMember(ctx, orgID, &request)
	if err != nil {
		if errors.Is(err, exception.ErrBadRequest) {
			response.WriteHTTPError(w, exception.ErrBadRequest, "invalid or non-existent organization role")
			return
		}
		if errors.Is(err, exception.ErrNotFound) {
			response.WriteHTTPError(w, exception.ErrNotFound, "organization not found")
			return
		}
		if errors.Is(err, exception.ErrForbidden) {
			response.WriteHTTPError(w, exception.ErrForbidden, "you do not have permission to manage organization members")
			return
		}
		if errors.Is(err, exception.ErrConflict) {
			response.WriteError(w, http.StatusConflict, err, "user is already a member of this organization")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to invite member")
		response.WriteHTTPError(w, err, "failed to invite member")
		return
	}
	response.WriteSuccess(w, http.StatusCreated, result)
}

// HTTPGetMembers retrieves all members of an organization
// @Summary      Get organization members
// @Description  Retrieves all members of an organization
// @Tags         organizations
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=[]model.MemberResponse}
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      403  {object}  response.SwaggerErrorResponseWrapper  "Forbidden"
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper  "Organization not found"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/{id}/members [get]
func (ctrl *OrganizationController) HTTPGetMembers(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	userID, _ := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	ctx := usecase.WithActorUserID(r.Context(), userID)
	result, err := ctrl.MemberUseCase.GetMembers(ctx, orgID)
	if err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			response.WriteHTTPError(w, exception.ErrNotFound, "organization not found")
			return
		}
		if errors.Is(err, exception.ErrForbidden) {
			response.WriteHTTPError(w, exception.ErrForbidden, "you do not have permission to view organization members")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to get members")
		response.WriteHTTPError(w, err, "failed to get members")
		return
	}
	response.WriteSuccess(w, http.StatusOK, result)
}

// HTTPUpdateMemberRole updates a member's role in an organization
// @Summary      Update member role
// @Description  Updates a member's role or status in an organization
// @Tags         organizations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                      true  "Organization ID"
// @Param        userId   path      string                      true  "User ID"
// @Param        request  body      model.UpdateMemberRequest   true  "Update member request"
// @Success      200      {object}  response.SwaggerSuccessResponseWrapper{data=model.MemberResponse}
// @Failure      400      {object}  response.SwaggerErrorResponseWrapper  "Validation error"
// @Failure      401      {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      403      {object}  response.SwaggerErrorResponseWrapper  "Forbidden"
// @Failure      404      {object}  response.SwaggerErrorResponseWrapper  "Member not found"
// @Failure      500      {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/{id}/members/{userId} [patch]
func (ctrl *OrganizationController) HTTPUpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	userID := r.PathValue("userId")
	actorUserID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || actorUserID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	var request model.UpdateMemberRequest
	if err := response.DecodeJSON(r, &request, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := ctrl.validate.Struct(&request); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}
	ctx := usecase.WithActorUserID(r.Context(), actorUserID)
	result, err := ctrl.MemberUseCase.UpdateMember(ctx, orgID, userID, &request)
	if err != nil {
		if errors.Is(err, exception.ErrBadRequest) {
			response.WriteHTTPError(w, exception.ErrBadRequest, "invalid or non-existent organization role")
			return
		}
		if errors.Is(err, exception.ErrNotFound) {
			response.WriteHTTPError(w, exception.ErrNotFound, "member not found")
			return
		}
		if errors.Is(err, exception.ErrForbidden) {
			response.WriteHTTPError(w, exception.ErrForbidden, "you do not have permission to update this member")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to update member")
		response.WriteHTTPError(w, err, "failed to update member")
		return
	}
	response.WriteSuccess(w, http.StatusOK, result)
}

// HTTPRemoveMember removes a member from an organization
// @Summary      Remove member from organization
// @Description  Removes a member from an organization (owner cannot be removed)
// @Tags         organizations
// @Security     BearerAuth
// @Produce      json
// @Param        id       path      string  true  "Organization ID"
// @Param        userId   path      string  true  "User ID"
// @Success      200      {object}  response.SwaggerSuccessResponseWrapper
// @Failure      401      {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      403      {object}  response.SwaggerErrorResponseWrapper  "Forbidden - cannot remove owner"
// @Failure      404      {object}  response.SwaggerErrorResponseWrapper  "Member not found"
// @Failure      500      {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/{id}/members/{userId} [delete]
func (ctrl *OrganizationController) HTTPRemoveMember(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	userID := r.PathValue("userId")
	actorUserID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || actorUserID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	ctx := usecase.WithActorUserID(r.Context(), actorUserID)
	err := ctrl.MemberUseCase.RemoveMember(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			response.WriteHTTPError(w, exception.ErrNotFound, "member not found")
			return
		}
		if errors.Is(err, exception.ErrForbidden) {
			response.WriteHTTPError(w, exception.ErrForbidden, "you do not have permission to remove this member")
			return
		}
		ctrl.Log.WithError(err).Error("Failed to remove member")
		response.WriteHTTPError(w, err, "failed to remove member")
		return
	}
	response.WriteSuccess(w, http.StatusOK, nil)
}

// HTTPGetPresence retrieves online members of an organization
// @Summary      Get online members
// @Description  Retrieves list of members who are currently online via WebSocket
// @Tags         organizations
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=[]interface{}}
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper  "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper  "Internal server error"
// @Router       /organizations/{id}/presence [get]
func (ctrl *OrganizationController) HTTPGetPresence(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		orgID = database.GetOrganizationID(r.Context())
	}
	result, err := ctrl.MemberUseCase.GetPresence(r.Context(), orgID)
	if err != nil {
		ctrl.Log.WithError(err).Error("Failed to get presence")
		response.WriteHTTPError(w, err, "failed to get presence")
		return
	}
	response.WriteSuccess(w, http.StatusOK, result)
}
