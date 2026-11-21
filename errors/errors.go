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
	err     error
	code    int
	details map[string]any
}

func New(text string, opts ...Option) Error {
	e := Error{err: errors.New(text)}
	for _, o := range opts {
		o(&e)
	}
	return e
}

func Wrap(err error, opts ...Option) Error {
	if err == nil {
		return Error{}
	}

	e := Error{err: err}
	var inner Error
	if errors.As(err, &inner) {
		e.code = inner.code
		e.details = copyDetails(inner.details)
	}
	for _, o := range opts {
		o(&e)
	}
	return e
}

func (e Error) Code() int {
	return e.code
}

func (e Error) Details() map[string]any {
	return copyDetails(e.details)
}

func (e Error) Error() string {
	if e.err == nil {
		return ""
	}

	return e.err.Error()
}

func (e Error) Unwrap() error {
	return e.err
}

func (e Error) Is(target error) bool {
	t, ok := target.(Error)
	if !ok {
		return false
	}

	return e.err == t.err && e.code == t.code
}

func (e Error) WriteResponse(rw http.ResponseWriter) {
	ngamux.Res(rw).Status(e.code).Text(e.Error())
}

func (e Error) MarshalJSON() ([]byte, error) {
	res := ngamux.Map{
		"error": e.Error(),
	}
	if e.code != 0 {
		res["code"] = e.code
	}
	if len(e.details) > 0 {
		res["details"] = e.details
	}
	return json.Marshal(res)
}

func IsCode(err error, code int) bool {
	var e Error
	return errors.As(err, &e) && e.Code() == code
}

type Option func(*Error)

func WithCode(code int) Option {
	return func(err *Error) {
		err.code = code
	}
}

func WithDetails(details map[string]any) Option {
	return func(err *Error) {
		err.details = copyDetails(details)
	}
}

func WithField(key string, value any, more ...any) Option {
	return func(err *Error) {
		if err.details == nil {
			err.details = make(map[string]any)
		}

		err.details[key] = value
		for i := 0; i+1 < len(more); i += 2 {
			fieldKey, ok := more[i].(string)
			if !ok {
				continue
			}
			err.details[fieldKey] = more[i+1]
		}
	}
}

func copyDetails(details map[string]any) map[string]any {
	if len(details) == 0 {
		return nil
	}

	result := make(map[string]any, len(details))
	for key, value := range details {
		result[key] = value
	}

	return result
}
