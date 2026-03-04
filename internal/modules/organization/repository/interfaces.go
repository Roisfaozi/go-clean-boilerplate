package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
)

// OrganizationRepository defines the interface for organization data access.
// All tenant-scoped queries should use the organization scope from context.
type OrganizationRepository interface {
	// Create creates a new organization with the owner as the first member.
	// This is an atomic operation that creates org + owner membership in a transaction.
	Create(ctx context.Context, org *entity.Organization, ownerRoleID string) error

	// FindByID finds an organization by its ID.
	FindByID(ctx context.Context, id string) (*entity.Organization, error)

	// FindBySlug finds an organization by its unique slug.
	FindBySlug(ctx context.Context, slug string) (*entity.Organization, error)

	// SlugExists checks if a slug is already taken.
	SlugExists(ctx context.Context, slug string) (bool, error)

	// FindUserOrganizations finds all organizations a user is a member of.
	FindUserOrganizations(ctx context.Context, userID string) ([]*entity.Organization, error)

	// Update updates an organization's details.
	Update(ctx context.Context, org *entity.Organization) error

	// Delete soft-deletes an organization.
	Delete(ctx context.Context, id string) error
}

// OrganizationMemberRepository defines the interface for membership data access.
type OrganizationMemberRepository interface {
	// CheckMembership verifies if a user is an active member of an organization.
	CheckMembership(ctx context.Context, orgID, userID string) (bool, error)

	// GetMemberStatus returns the membership status of a user in an organization.
	GetMemberStatus(ctx context.Context, orgID, userID string) (string, error)

	// AddMember adds a user to an organization with a specific role.
	AddMember(ctx context.Context, member *entity.OrganizationMember) error

	// RemoveMember removes a user from an organization.
	RemoveMember(ctx context.Context, orgID, userID string) error

	// UpdateMemberRole updates a member's role in an organization.
	UpdateMemberRole(ctx context.Context, orgID, userID, roleID string) error

	// UpdateMemberStatus updates a member's status (active, suspended, banned).
	UpdateMemberStatus(ctx context.Context, orgID, userID, status string) error

	// FindMembers finds all members of an organization.
	FindMembers(ctx context.Context, orgID string) ([]*entity.OrganizationMember, error)

	// GetMemberRole returns the role of a user in an organization.
	GetMemberRole(ctx context.Context, orgID, userID string) (string, error)
}
