package usecase

type IEnforcer interface {
	AddGroupingPolicy(params ...interface{}) (bool, error)
	AddPolicy(params ...interface{}) (bool, error)
	RemovePolicy(params ...interface{}) (bool, error)
	GetPolicy() ([][]string, error)
	GetFilteredPolicy(fieldIndex int, fieldValues ...string) ([][]string, error)
	UpdatePolicy(oldRule []string, newRule []string) (bool, error)
	GetRolesForUser(name string, domain ...string) ([]string, error)
	RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) (bool, error)
	Enforce(params ...interface{}) (bool, error)
}
