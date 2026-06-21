package usecase

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
)

// OrganizationUseCase defines the interface for organization business logic
type OrganizationUseCase interface {
	// CreateOrganization creates a new organization with the current user as owner
	CreateOrganization(ctx context.Context, userID string, request *model.CreateOrganizationRequest) (*model.OrganizationResponse, error)

	// GetOrganization retrieves an organization by ID
	GetOrganization(ctx context.Context, id string) (*model.OrganizationResponse, error)

	// GetOrganizationBySlug retrieves an organization by slug
	GetOrganizationBySlug(ctx context.Context, slug string) (*model.OrganizationResponse, error)

	// UpdateOrganization updates an organization's details
	UpdateOrganization(ctx context.Context, id string, request *model.UpdateOrganizationRequest) (*model.OrganizationResponse, error)

	// GetUserOrganizations retrieves all organizations a user is a member of
	GetUserOrganizations(ctx context.Context, userID string) (*model.UserOrganizationsResponse, error)

	// DeleteOrganization deletes an organization (owner only)
	DeleteOrganization(ctx context.Context, id string, userID string) error
}

// OrganizationMemberUseCase defines the interface for member management business logic
type OrganizationMemberUseCase interface {
	// InviteMember invites a user to an organization
	InviteMember(ctx context.Context, orgID string, request *model.InviteMemberRequest) (*model.MemberResponse, error)

	// GetMembers retrieves all members of an organization
	GetMembers(ctx context.Context, orgID string) ([]model.MemberResponse, error)

	// UpdateMember updates a member's role or status
	UpdateMember(ctx context.Context, orgID, userID string, request *model.UpdateMemberRequest) (*model.MemberResponse, error)

	// RemoveMember removes a member from an organization
	RemoveMember(ctx context.Context, orgID, userID string) error

	// AcceptInvitation accepts an invitation
	AcceptInvitation(ctx context.Context, request *model.AcceptInvitationRequest) error

	// GetPresence retrieves online members of an organization
	GetPresence(ctx context.Context, orgID string) ([]interface{}, error)
}

// IOrganizationReader provides high-performance membership validation with caching.
// Used by TenantMiddleware to avoid direct database hits on every request.
type IOrganizationReader interface {
	// ValidateMembership checks if a user is an active member of an organization.
	// Uses Redis cache for performance, falls back to database on cache miss.
	ValidateMembership(ctx context.Context, orgID, userID string) (bool, error)

	// GetMemberRole returns the role of a user in an organization (owner, admin, member).
	// Returns empty string if not a member.
	GetMemberRole(ctx context.Context, orgID, userID string) (string, error)

	// InvalidateMembershipCache removes the cached membership data for a user-org pair.
	// Must be called when membership changes (add, remove, role update).
	InvalidateMembershipCache(ctx context.Context, orgID, userID string) error

	// InvalidateOrganizationCache removes all cached membership data for an organization.
	// Must be called when organization is deleted.
	InvalidateOrganizationCache(ctx context.Context, orgID string) error
}
