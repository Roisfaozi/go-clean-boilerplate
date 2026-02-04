package converter

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	userModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
)

// OrganizationToResponse converts an Organization entity to a response DTO
func OrganizationToResponse(org *entity.Organization) *model.OrganizationResponse {
	if org == nil {
		return nil
	}
	return &model.OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		OwnerID:   org.OwnerID,
		Status:    org.Status,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}
}

// OrganizationsToResponse converts a slice of Organization entities to response DTOs
func OrganizationsToResponse(orgs []*entity.Organization) []model.OrganizationResponse {
	responses := make([]model.OrganizationResponse, 0, len(orgs))
	for _, org := range orgs {
		if resp := OrganizationToResponse(org); resp != nil {
			responses = append(responses, *resp)
		}
	}
	return responses
}

// MemberToResponse converts an OrganizationMember entity to a response DTO
func MemberToResponse(member *entity.OrganizationMember) *model.MemberResponse {
	if member == nil {
		return nil
	}

	resp := &model.MemberResponse{
		ID:             member.ID,
		OrganizationID: member.OrganizationID,
		UserID:         member.UserID,
		RoleID:         member.RoleID,
		Status:         member.Status,
		JoinedAt:       member.JoinedAt,
	}

	if member.User.ID != "" {
		resp.User = &userModel.UserResponse{
			ID:        member.User.ID,
			Email:     member.User.Email,
			Username:  member.User.Username,
			Name:      member.User.Name,
			AvatarURL: member.User.AvatarURL,
			Status:    member.User.Status,
			CreatedAt: member.User.CreatedAt,
			UpdatedAt: member.User.UpdatedAt,
		}
	}

	return resp
}

// MembersToResponse converts a slice of OrganizationMember entities to response DTOs
func MembersToResponse(members []*entity.OrganizationMember) []model.MemberResponse {
	responses := make([]model.MemberResponse, 0, len(members))
	for _, member := range members {
		if resp := MemberToResponse(member); resp != nil {
			responses = append(responses, *resp)
		}
	}
	return responses
}
