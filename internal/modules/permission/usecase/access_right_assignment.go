package usecase

import (
	"context"
	"fmt"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
)

func (uc *PermissionUseCase) GetRoleAccessRights(ctx context.Context, role, domain string) ([]model.RoleAccessRightStatus, error) {
	if domain == "" {
		domain = "global"
	}

	accessRights, err := uc.AccessRepo.GetAccessRights(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch access rights: %w", err)
	}

	enf := uc.enforcer.WithContext(ctx)

	result := make([]model.RoleAccessRightStatus, 0, len(accessRights))
	for _, ar := range accessRights {
		granted := 0
		endpointLabels := make([]string, 0, len(ar.Endpoints))

		for _, ep := range ar.Endpoints {
			endpointLabels = append(endpointLabels, fmt.Sprintf("%s %s", ep.Method, ep.Path))
			ok, _ := enf.Enforce(role, domain, ep.Path, ep.Method)
			if ok {
				granted++
			}
		}

		total := len(ar.Endpoints)
		status := model.RoleAccessRightStatus{
			ID:        ar.ID,
			Name:      ar.Name,
			Endpoints: endpointLabels,
			Assigned:  total > 0 && granted == total,
			Partial:   granted > 0 && granted < total,
		}
		result = append(result, status)
	}

	return result, nil
}

func (uc *PermissionUseCase) AssignAccessRight(ctx context.Context, req model.AssignAccessRightRequest) error {
	if req.Domain == "" {
		req.Domain = "global"
	}

	ar, err := uc.AccessRepo.GetAccessRightByID(ctx, req.AccessRightID)
	if err != nil {
		return exception.ErrNotFound
	}
	if len(ar.Endpoints) == 0 {
		return fmt.Errorf("access right '%s' has no endpoints configured", ar.Name)
	}

	enf := uc.enforcer.WithContext(ctx)
	for _, ep := range ar.Endpoints {
		if ok, _ := enf.Enforce(req.Role, req.Domain, ep.Path, ep.Method); !ok {
			if _, err := enf.AddPolicy(req.Role, req.Domain, ep.Path, ep.Method); err != nil {
				return fmt.Errorf("failed to grant %s %s to %s: %w", ep.Method, ep.Path, req.Role, err)
			}
		}
	}

	uc.log.Infof("Assigned access right '%s' (%d endpoints) to role '%s' in domain '%s'",
		ar.Name, len(ar.Endpoints), req.Role, req.Domain)

	if uc.AuditUC != nil {
		_ = uc.AuditUC.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			Action:    "ASSIGN_ACCESS_RIGHT",
			Entity:    "roles",
			EntityID:  req.Role,
			NewValues: map[string]any{"access_right_id": ar.ID, "domain": req.Domain},
		})
	}

	return nil
}

func (uc *PermissionUseCase) RevokeAccessRight(ctx context.Context, req model.AssignAccessRightRequest) error {
	if req.Domain == "" {
		req.Domain = "global"
	}

	ar, err := uc.AccessRepo.GetAccessRightByID(ctx, req.AccessRightID)
	if err != nil {
		return exception.ErrNotFound
	}

	enf := uc.enforcer.WithContext(ctx)
	for _, ep := range ar.Endpoints {
		if _, err := enf.RemovePolicy(req.Role, req.Domain, ep.Path, ep.Method); err != nil {
			return fmt.Errorf("failed to revoke %s %s from %s: %w", ep.Method, ep.Path, req.Role, err)
		}
	}

	uc.log.Infof("Revoked access right '%s' (%d endpoints) from role '%s' in domain '%s'",
		ar.Name, len(ar.Endpoints), req.Role, req.Domain)

	if uc.AuditUC != nil {
		_ = uc.AuditUC.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			Action:    "REVOKE_ACCESS_RIGHT",
			Entity:    "roles",
			EntityID:  req.Role,
			OldValues: map[string]any{"access_right_id": ar.ID, "domain": req.Domain},
		})
	}

	return nil
}
