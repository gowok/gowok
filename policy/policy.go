package policy

import (
	"github.com/casbin/casbin/v2"
	cModel "github.com/casbin/casbin/v2/model"
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

func NewPolicy(model string) (*Policy, error) {
	m, err := cModel.NewModelFromString(model)
	if err != nil {
		return nil, err
	}

	e, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, err
	}

	ee := &Policy{e}
	return ee, nil
}

func NewPolicyRBAC() (*Policy, error) {
	model := `
	[request_definition]
	r = sub, obj, act

	[policy_definition]
	p = sub, obj, act

	[role_definition]
	g = _, _

	[policy_effect]
	e = some(where (p.eft == allow))

	[matchers]
  m = (g(r.sub, p.sub) || r.sub == p.sub) && r.obj == p.obj && r.act == p.act
	`

	return NewPolicy(model)
}

func NewPolicyABAC() (*Policy, error) {
	model := `
  [request_definition]
  r = sub, obj, act

  [policy_definition]
  p = sub, obj, act

  [role_definition]
  g = _, _

  [policy_effect]
  e = some(where (p.eft == allow))

  [matchers]
  m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
	`

	return NewPolicy(model)
}
