package gowok

import (
	"testing"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/gowok/should"
)

func validatorInitialize(t *testing.T) *Validator {
	val := NewValidator()
	should.NotNil(t, val)

	en := en.New()
	uni := ut.New(en, en)
	trans, ok := uni.GetTranslator("en")
	should.True(t, ok)
	should.NotNil(t, trans)
	val.SetTranslator(trans, func(v *validator.Validate, trans ut.Translator) error {
		return en_translations.RegisterDefaultTranslations(val.validate, trans)
	})

	should.NotNil(t, val.trans)

	return val
}

func TestNewValidator(t *testing.T) {
	validatorInitialize(t)
}

func TestValidateStruct(t *testing.T) {
	type user struct {
		Email string `validate:"required"`
	}

	val := validatorInitialize(t)
	tests := []struct {
		name            string
		input           user
		inputTrans      map[string]string
		expectedIsError bool
		expectedMessage string
	}{
		{
			"positive",
			user{"user@mail.com"},
			map[string]string{},
			false,
			"",
		},
		{
			"negative",
			user{},
			map[string]string{},
			true,
			"",
		},
		{
			"negative with trans",
			user{},
			map[string]string{
				"required": "err",
			},
			true,
			"",
		},
		{
			"negative with trans per field",
			user{},
			map[string]string{
				"user.Email.required": "err",
			},
			true,
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := val.ValidateStruct(test.input, test.inputTrans)
			should.Equal(t, len(err.Error()) > 0, test.expectedIsError)
		})
	}
}
