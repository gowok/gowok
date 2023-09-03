package exception

import (
	"errors"
	"fmt"
)

var (
	ConfigNotFound = errors.New("config file not found")
	ConfigDecoding = func(err error) error { return fmt.Errorf("config decoding failed: %s", err.Error()) }

	EmailAlreadyUsed       = errors.New("email already used")
	EmailOrPasswordInvalid = errors.New("email or password invalid")
	InvalidCredentials     = errors.New("invalid credentials")

	TokenGenerationFail = errors.New("token generation fail")
	InvalidPolicyData   = errors.New("invalid policy data")
	DeleteRoleWithUsers = errors.New("can't delete role with some users")

	ErrNilPointerDeref = errors.New("nil pointer dereference")
	ErrGetOfNoValue    = errors.New("get of no value")
	ErrNoDatabaseFound = errors.New("no database found")
)
