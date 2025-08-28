package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrConfigNotFound = errors.New("config file not found")
	ErrConfigDecoding = func(err error) error { return fmt.Errorf("config decoding failed: %s", err.Error()) }
	ErrNotConfigured  = func(name string) error { return fmt.Errorf("%s not configured", name) }

	ErrEmailAlreadyUsed       = errors.New("email already used")
	ErrEmailOrPasswordInvalid = errors.New("email or password invalid")
	ErrInvalidCredentials     = errors.New("invalid credentials")

	ErrTokenGenerationFail = errors.New("token generation fail")
	ErrInvalidPolicyData   = errors.New("invalid policy data")
	ErrDeleteRoleWithUsers = errors.New("can't delete role with some users")

	ErrNilPointerDeref = errors.New("nil pointer dereference")
	ErrGetOfNoValue    = errors.New("get of no value")
	ErrNoDatabaseFound = errors.New("no database found")

	ErrNoValuePresent = errors.New("no value present")
)
