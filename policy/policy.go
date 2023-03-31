package policy

import (
	"github.com/casbin/casbin/v2"
)

const (
	ActionCreate = "CREATE"
	ActionRead   = "READ"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
)

var Actions = []string{
	ActionCreate,
	ActionRead,
	ActionUpdate,
	ActionDelete,
}

type Policy struct {
	*casbin.Enforcer
}

func NewPolicy(params ...interface{}) (*Policy, error) {
	e, err := casbin.NewEnforcer(params...)
	if err != nil {
		return nil, err
	}

	ee := &Policy{e}
	return ee, nil
}
