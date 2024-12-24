package gowok

import (
	"encoding/json"
	"net/http"
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/ngamux/ngamux"
)

// Validator struct
type Validator struct {
	validate *validator.Validate
	trans    ut.Translator
}

// ValidationError represent error from validator
type ValidationError struct {
	errors       validator.ValidationErrors
	translations validator.ValidationErrorsTranslations

	errorMessage string
	errorJSON    map[string]string
}

func NewValidationError(errs validator.ValidationErrors, translations ut.Translator, trans map[string]string) ValidationError {
	validationErrorMessages := errs.Translate(translations)
	errorMessage := ""
	errorJSON := map[string]string{}
	for _, err := range errs {
		namespace := err.Namespace()
		errorJSON[namespace] = validationErrorMessages[namespace]

		customTag := namespace + "." + err.Tag()
		if _, ok := trans[customTag]; ok {
			t, e := translations.T(customTag, err.Field())
			if e == nil {
				errorJSON[namespace] = t
			}
			continue
		}

		customTag = "*." + err.Tag()
		if _, ok := trans[err.Tag()]; ok {
			t, e := translations.T(customTag, err.Field())
			if e == nil {
				errorJSON[namespace] = t
			}
		}
	}

	return ValidationError{errs, validationErrorMessages, errorMessage, errorJSON}
}

func (err ValidationError) Error() string {
	errorMessage := ""
	for field, msg := range err.errorJSON {
		errorMessage += field + ": " + msg + "; "
	}
	return errorMessage
}

func (err ValidationError) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.errorJSON)
}

// NewValidator create an instance of Validator Struct
func NewValidator() *Validator {
	validate := validator.New()

	_validator := &Validator{
		validate: validate,
		trans:    nil,
	}

	return _validator
}

func (v *Validator) SetTranslator(trans ut.Translator, localeFunc func(*validator.Validate, ut.Translator) error) error {
	v.trans = trans
	return localeFunc(v.validate, trans)
}

// ValidateStruct func
func (v *Validator) ValidateStruct(input any, trans map[string]string) ValidationError {
	for tag, message := range trans {
		if !strings.Contains(tag, ".") {
			tag = "*." + tag
		}
		err := v.registerTranslationTag(tag, message, false)
		if err != nil {
			return ValidationError{
				errorJSON: map[string]string{"*": err.Error()},
			}
		}
	}

	err := v.validate.Struct(input)
	if err != nil {
		switch e := err.(type) {
		case validator.ValidationErrors:
			errResp := NewValidationError(e, v.trans, trans)
			return errResp
		}
	}

	return ValidationError{}
}

func (v *Validator) registerTranslationTag(tag, message string, override bool) error {
	err := v.validate.RegisterTranslation(tag, v.trans, func(ut ut.Translator) error {
		return ut.Add(tag, message, override)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Field())
		return t
	})
	return err
}

func ValidateJSON[T any](r *http.Request, schema T, trans map[string]string) (*T, *ValidationError) {
	project := Get()
	err := ngamux.Req(r).JSON(&schema)
	if err != nil {
		return new(T), &ValidationError{
			errorMessage: err.Error(),
			errorJSON: map[string]string{
				"*": err.Error(),
			},
		}
	}

	errs := project.Validator.ValidateStruct(schema, trans)
	if len(errs.Error()) > 0 {
		return new(T), &errs
	}

	return &schema, nil
}
