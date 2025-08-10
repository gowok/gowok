package exception

import (
	"errors"
	"fmt"
)

var (
	NotImplemented = errors.New("not implemented")
	ConfigNotFound = errors.New("config file not found")
	ConfigDecoding = func(err error) error { return fmt.Errorf("config decoding failed: %s", err.Error()) }

	EmailAlreadyUsed       = errors.New("email already used")
	EmailOrPasswordInvalid = errors.New("email or password invalid")
	InvalidCredentials     = errors.New("invalid credentials")

	TokenGenerationFail = errors.New("token generation fail")
	InvalidPolicyData   = errors.New("invalid policy data")
	DeleteRoleWithUsers = errors.New("can't delete role with some users")

	NilPointerDeref = errors.New("nil pointer dereference")
	GetOfNoValue    = errors.New("get of no value")
	NoDatabaseFound = errors.New("no database found")

	NoValuePresent = errors.New("no value present")
)
