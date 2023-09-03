package gowok

import (
	"encoding/json"
	"fmt"

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
	errors       validator.ValidationErrors
	translations validator.ValidationErrorsTranslations

	errorMessage string
	errorJSON    map[string]string
}

func NewValidationError(errs validator.ValidationErrors, translations ut.Translator) ValidationError {
	validationErrorMessages := errs.Translate(translations)
	errorMessage := ""
	errorJSON := map[string]string{}
	for _, err := range errs {
		field := err.Field()
		namespace := err.Namespace()
		errorMessage += fmt.Sprintf("%s: %s; ", field, validationErrorMessages[namespace])
		errorJSON[field] = validationErrorMessages[namespace]
	}

	return ValidationError{errs, validationErrorMessages, errorMessage, errorJSON}
}

func (err ValidationError) Error() string {
	return err.errorMessage
}

func (err ValidationError) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.errorJSON)
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

// ValidateStruct func
func (v *Validator) ValidateStruct(input any) ValidationError {
	err := v.validate.Struct(input)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errResp := NewValidationError(validationErrors, v.trans)
		return errResp
	}

	return ValidationError{}
}
