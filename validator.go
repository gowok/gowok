package gowok

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// Validator struct
type Validator struct {
	validate *validator.Validate
	trans    ut.Translator
}

// ValidationError represent error from validator
type ValidationError struct {
	Namespace string `json:"namespace,omitempty"`
	Field     string `json:"field,omitempty"`
	Error     string `json:"error,omitempty"`
}

// NewValidator create an instance of Validator Struct
func NewValidator() *Validator {
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	validate := validator.New()
	en_translations.RegisterDefaultTranslations(validate, trans)

	_validator := &Validator{
		validate: validate,
		trans:    trans,
	}

	return _validator
}

func (v *Validator) formatErrors(errs validator.ValidationErrors) []ValidationError {
	validationErrorMessages := errs.Translate(v.trans)

	messages := make([]ValidationError, 0)
	for _, err := range errs {
		messages = append(messages, ValidationError{
			Namespace: err.Namespace(),
			Field:     err.Field(),
			Error:     validationErrorMessages[err.Namespace()],
		})
	}

	return messages
}

// ValidateStruct func
func (v *Validator) ValidateStruct(input any) []ValidationError {
	err := v.validate.Struct(input)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errResp := v.formatErrors(validationErrors)
		return errResp
	}

	return []ValidationError{}
}
