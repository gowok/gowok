package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gowok/gowok/json"
	"github.com/ngamux/ngamux"
)

var (
	ErrNotImplemented       = New("not implemented")
	ErrConfigNotFound       = New("config file not found")
	ErrStreamingUnsupported = New("streaming unsupported")

	ErrConfigDecoding = func(err error) error { return New(fmt.Sprintf("config decoding failed: %s", err.Error())) }
	ErrNotConfigured  = func(name string) error { return New(fmt.Sprintf("%s not configured", name)) }

	ErrEmailAlreadyUsed       = New("email already used")
	ErrEmailOrPasswordInvalid = New("email or password invalid")
	ErrInvalidCredentials     = New("invalid credentials")

	ErrTokenGenerationFail = New("token generation fail")
	ErrInvalidPolicyData   = New("invalid policy data")
	ErrDeleteRoleWithUsers = New("can't delete role with some users")

	ErrNilPointerDeref = New("nil pointer dereference")
	ErrGetOfNoValue    = New("get of no value")
	ErrNoDatabaseFound = New("no database found")

	ErrNoValuePresent = New("no value present")
)

type Error struct {
	err  error
	code int
}

func New(text string, opts ...Option) Error {
	e := Error{errors.New(text), 0}
	for _, o := range opts {
		o(&e)
	}
	return e
}

func (e Error) Error() string {
	return e.err.Error()
}

func (e Error) WriteResponse(rw http.ResponseWriter) {
	ngamux.Res(rw).Status(e.code).Text(e.Error())
}

func (e Error) MarshalJSON() ([]byte, error) {
	res := ngamux.Map{
		"error": e.err.Error(),
	}
	if e.code != 0 {
		res["code"] = e.code
	}
	return json.Marshal(res)
}

type Option func(*Error)

func WithCode(code int) Option {
	return func(err *Error) {
		err.code = code
	}
}
