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
	adapter any
}

func NewPolicy(model string, opts ...Option) (*Policy, error) {
	m, err := cModel.NewModelFromString(model)
	if err != nil {
		return nil, err
	}

	ee := &Policy{}
	for _, opt := range opts {
		opt(ee)
	}

	params := make([]any, 0)
	params = append(params, m)
	if ee.adapter != nil {
		params = append(params, ee.adapter)
	}

	e, err := casbin.NewEnforcer(params...)
	if err != nil {
		return nil, err
	}
	ee.Enforcer = e

	return ee, nil
}

func NewPolicyRBAC(opts ...Option) (*Policy, error) {
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

	return NewPolicy(model, opts...)
}

func NewPolicyABAC(opts ...Option) (*Policy, error) {
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

	return NewPolicy(model, opts...)
}
