package middleware

import (
	"errors"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
)

// HTTPRequireOrganization enforces organization presence and tenant isolation for net/http.
func (m *TenantMiddleware) HTTPRequireOrganization() delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
			if !ok {
				response.WriteError(w, http.StatusUnauthorized, errors.New("user not authenticated"), "unauthorized")
				return
			}

			allowDeleted := database.CanAccessDeletedOrganizations(r.Context())
			orgID := r.Header.Get(OrgIDHeader)
			orgSlug := r.Header.Get(OrgSlugHeader)
			if orgID == "" {
				orgID, _ = delivery.GetContextString(r.Context(), delivery.OrganizationIDKey)
			}
			if orgID == "" {
				if id := r.PathValue("id"); id != "" {
					if id == "slug" {
						if action := r.PathValue("action"); action != "" && orgSlug == "" {
							orgSlug = action
						}
					} else {
						orgID = id
					}
				}
			}
			if orgSlug == "" {
				if slug := r.PathValue("slug"); slug != "" {
					orgSlug = slug
				}
			}

			if orgID == "" && orgSlug == "" {
				response.WriteError(w, http.StatusBadRequest, errors.New("organization ID or slug is required"), "missing organization identifier")
				return
			}

			if orgID == "" && orgSlug != "" {
				org, err := m.OrgRepo.FindBySlug(r.Context(), orgSlug)
				if err != nil || org == nil {
					response.WriteError(w, http.StatusNotFound, errors.New("organization not found"), "organization not found")
					return
				}
				orgID = org.ID
			}

			isMember, err := m.Reader.ValidateMembership(r.Context(), orgID, userID)
			if err != nil {
				if m.Log != nil {
					m.Log.WithError(err).Error("Failed to validate membership")
				}
				response.WriteError(w, http.StatusInternalServerError, err, "internal server error")
				return
			}

			if !isMember && !allowDeleted {
				response.WriteError(w, http.StatusForbidden, errors.New("user is not a member of this organization"), "access denied")
				return
			}

			ctx := database.SetOrganizationContext(r.Context(), orgID)
			ctx = delivery.SetContextValue(ctx, delivery.OrganizationIDKey, orgID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// HTTPOptionalOrganization injects organization context if present in header without requiring it.
func (m *TenantMiddleware) HTTPOptionalOrganization() delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID := r.Header.Get(OrgIDHeader)
			orgSlug := r.Header.Get(OrgSlugHeader)

			if orgID == "" && orgSlug != "" {
				org, err := m.OrgRepo.FindBySlug(r.Context(), orgSlug)
				if err == nil && org != nil {
					orgID = org.ID
				}
			}

			if orgID != "" {
				ctx := database.SetOrganizationContext(r.Context(), orgID)
				ctx = delivery.SetContextValue(ctx, delivery.OrganizationIDKey, orgID)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}
