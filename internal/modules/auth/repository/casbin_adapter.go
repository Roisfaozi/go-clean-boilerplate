package repository

import (
	"context"

	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
)

type casbinAdapter struct {
	enforcer      permissionUseCase.IEnforcer
	defaultRole   string
	defaultDomain string
}

const (
	defaultCasbinRole   = "role:user"
	defaultCasbinDomain = "global"
)

func NewCasbinAdapter(enforcer permissionUseCase.IEnforcer, defaultRole, defaultDomain string) AuthzManager {
	if defaultRole == "" {
		defaultRole = defaultCasbinRole
	}
	if defaultDomain == "" {
		defaultDomain = defaultCasbinDomain
	}

	return &casbinAdapter{
		enforcer:      enforcer,
		defaultRole:   defaultRole,
		defaultDomain: defaultDomain,
	}
}

func (a *casbinAdapter) AssignDefaultRole(ctx context.Context, userID string) error {
	if a.enforcer == nil {
		return nil
	}
	_, err := a.enforcer.WithContext(ctx).AddGroupingPolicy(userID, a.defaultRole, a.defaultDomain)
	return err
}

func (a *casbinAdapter) GetRolesForUser(ctx context.Context, userID string, domain string) ([]string, error) {
	if a.enforcer == nil {
		return nil, nil
	}
	if domain == "" {
		domain = a.defaultDomain
	}
	return a.enforcer.GetRolesForUser(userID, domain)
}
