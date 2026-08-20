package usecase

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

type transactionalEnforcer struct {
	globalEnforcer *casbin.Enforcer
	casbinModel    string
}

func NewTransactionalEnforcer(globalEnforcer *casbin.Enforcer, casbinModelPath string) IEnforcer {
	globalEnforcer.EnableAutoSave(true)
	return &transactionalEnforcer{
		globalEnforcer: globalEnforcer,
		casbinModel:    casbinModelPath,
	}
}

func (e *transactionalEnforcer) getEnforcer(ctx context.Context) IEnforcer {
	if txDB, ok := tx.DBFromContext(ctx); ok {
		adapter, err := gormadapter.NewAdapterByDB(txDB)
		if err != nil {
			return &failingEnforcer{err: err}
		}

		enforcer, err := casbin.NewEnforcer(e.casbinModel, adapter)
		if err != nil {
			return &failingEnforcer{err: err}
		}
		enforcer.EnableAutoSave(true)

		return &transientEnforcer{inner: enforcer}
	}

	return e
}

func (e *transactionalEnforcer) WithContext(ctx context.Context) IEnforcer {
	return e.getEnforcer(ctx)
}

func (e *transactionalEnforcer) AddGroupingPolicy(params ...interface{}) (bool, error) {
	return e.globalEnforcer.AddGroupingPolicy(params...)
}

func (e *transactionalEnforcer) AddPolicy(params ...interface{}) (bool, error) {
	return e.globalEnforcer.AddPolicy(params...)
}

func (e *transactionalEnforcer) HasGroupingPolicy(params ...interface{}) (bool, error) {
	return e.globalEnforcer.HasGroupingPolicy(params...)
}

func (e *transactionalEnforcer) HasPolicy(params ...interface{}) (bool, error) {
	return e.globalEnforcer.HasPolicy(params...)
}

func (e *transactionalEnforcer) RemovePolicy(params ...interface{}) (bool, error) {
	return e.globalEnforcer.RemovePolicy(params...)
}

func (e *transactionalEnforcer) GetPolicy() ([][]string, error) {
	return e.globalEnforcer.GetPolicy()
}

func (e *transactionalEnforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) ([][]string, error) {
	return e.globalEnforcer.GetFilteredPolicy(fieldIndex, fieldValues...)
}

func (e *transactionalEnforcer) UpdatePolicy(oldRule []string, newRule []string) (bool, error) {
	return e.globalEnforcer.UpdatePolicy(oldRule, newRule)
}

func (e *transactionalEnforcer) GetRolesForUser(name string, domain ...string) ([]string, error) {
	return e.globalEnforcer.GetRolesForUser(name, domain...)
}

func (e *transactionalEnforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return e.globalEnforcer.RemoveFilteredGroupingPolicy(fieldIndex, fieldValues...)
}

func (e *transactionalEnforcer) Enforce(params ...interface{}) (bool, error) {
	return e.globalEnforcer.Enforce(params...)
}

func (e *transactionalEnforcer) GetUsersForRole(name string, domain ...string) ([]string, error) {
	return e.globalEnforcer.GetUsersForRole(name, domain...)
}

func (e *transactionalEnforcer) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return e.globalEnforcer.RemoveFilteredPolicy(fieldIndex, fieldValues...)
}

func (e *transactionalEnforcer) DeleteRole(role string) (bool, error) {
	// Built-in DeleteRole should work if AutoSave is on.
	// It removes from both policy (p) and grouping policy (g).
	return e.globalEnforcer.DeleteRole(role)
}

func (e *transactionalEnforcer) SavePolicy() error {
	return e.globalEnforcer.SavePolicy()
}

func (e *transactionalEnforcer) LoadPolicy() error {
	return e.globalEnforcer.LoadPolicy()
}

type transientEnforcer struct {
	inner *casbin.Enforcer
}

func (e *transientEnforcer) WithContext(ctx context.Context) IEnforcer {
	return e
}

func (e *transientEnforcer) AddGroupingPolicy(params ...interface{}) (bool, error) {
	return e.inner.AddGroupingPolicy(params...)
}

func (e *transientEnforcer) AddPolicy(params ...interface{}) (bool, error) {
	return e.inner.AddPolicy(params...)
}

func (e *transientEnforcer) HasGroupingPolicy(params ...interface{}) (bool, error) {
	return e.inner.HasGroupingPolicy(params...)
}

func (e *transientEnforcer) HasPolicy(params ...interface{}) (bool, error) {
	return e.inner.HasPolicy(params...)
}

func (e *transientEnforcer) RemovePolicy(params ...interface{}) (bool, error) {
	return e.inner.RemovePolicy(params...)
}

func (e *transientEnforcer) GetPolicy() ([][]string, error) {
	return e.inner.GetPolicy()
}

func (e *transientEnforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) ([][]string, error) {
	return e.inner.GetFilteredPolicy(fieldIndex, fieldValues...)
}

func (e *transientEnforcer) UpdatePolicy(oldRule []string, newRule []string) (bool, error) {
	return e.inner.UpdatePolicy(oldRule, newRule)
}

func (e *transientEnforcer) GetRolesForUser(name string, domain ...string) ([]string, error) {
	return e.inner.GetRolesForUser(name, domain...)
}

func (e *transientEnforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return e.inner.RemoveFilteredGroupingPolicy(fieldIndex, fieldValues...)
}

func (e *transientEnforcer) Enforce(params ...interface{}) (bool, error) {
	return e.inner.Enforce(params...)
}

func (e *transientEnforcer) GetUsersForRole(name string, domain ...string) ([]string, error) {
	return e.inner.GetUsersForRole(name, domain...)
}

func (e *transientEnforcer) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return e.inner.RemoveFilteredPolicy(fieldIndex, fieldValues...)
}

func (e *transientEnforcer) DeleteRole(role string) (bool, error) {
	return e.inner.DeleteRole(role)
}

func (e *transientEnforcer) SavePolicy() error {
	return e.inner.SavePolicy()
}

func (e *transientEnforcer) LoadPolicy() error {
	return e.inner.LoadPolicy()
}

type failingEnforcer struct {
	err error
}

func (e *failingEnforcer) WithContext(ctx context.Context) IEnforcer {
	return e
}

func (e *failingEnforcer) AddGroupingPolicy(params ...interface{}) (bool, error) {
	return false, e.err
}

func (e *failingEnforcer) AddPolicy(params ...interface{}) (bool, error) {
	return false, e.err
}

func (e *failingEnforcer) HasGroupingPolicy(params ...interface{}) (bool, error) {
	return false, e.err
}

func (e *failingEnforcer) HasPolicy(params ...interface{}) (bool, error) {
	return false, e.err
}

func (e *failingEnforcer) RemovePolicy(params ...interface{}) (bool, error) {
	return false, e.err
}

func (e *failingEnforcer) GetPolicy() ([][]string, error) {
	return nil, e.err
}

func (e *failingEnforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) ([][]string, error) {
	return nil, e.err
}

func (e *failingEnforcer) UpdatePolicy(oldRule []string, newRule []string) (bool, error) {
	return false, e.err
}

func (e *failingEnforcer) GetRolesForUser(name string, domain ...string) ([]string, error) {
	return nil, e.err
}

func (e *failingEnforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return false, e.err
}

func (e *failingEnforcer) Enforce(params ...interface{}) (bool, error) {
	return false, e.err
}

func (e *failingEnforcer) GetUsersForRole(name string, domain ...string) ([]string, error) {
	return nil, e.err
}

func (e *failingEnforcer) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return false, e.err
}

func (e *failingEnforcer) DeleteRole(role string) (bool, error) {
	return false, e.err
}

func (e *failingEnforcer) SavePolicy() error {
	return e.err
}

func (e *failingEnforcer) LoadPolicy() error {
	return e.err
}
