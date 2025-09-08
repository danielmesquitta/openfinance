package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

type Validator struct {
	val   *validator.Validate
	trans ut.Translator
}

func NewValidator() *Validator {
	validate := validator.New()
	english := en.New()
	uni := ut.New(english, english)
	trans, ok := uni.GetTranslator("en")
	if !ok {
		panic("translator not found")
	}

	if err := enTranslations.
		RegisterDefaultTranslations(validate, trans); err != nil {
		panic(err)
	}

	return &Validator{
		validate,
		trans,
	}
}

// Validate validates the data (struct)
// returning an error if the data is invalid.
func (v *Validator) Validate(
	data any,
) error {
	err := v.val.Struct(data)
	if err == nil {
		return nil
	}

	var validationErrs validator.ValidationErrors
	if ok := errors.As(err, &validationErrs); !ok {
		return fmt.Errorf("failed to validate data: %w", err)
	}

	strErrs := make([]string, len(validationErrs))
	for i, validationErr := range validationErrs {
		strErrs[i] = validationErr.Translate(v.trans)
	}

	errMsg := strings.Join(
		strErrs,
		", ",
	)

	return errors.New(errMsg)
}
