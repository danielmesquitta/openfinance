package validator

import (
	"fmt"
	"strings"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

type Validator struct {
	validate *validator.Validate
	trans    ut.Translator
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

func (v *Validator) Validate(
	data any,
) *entity.AppError {
	var strErrs []string

	err := v.validate.Struct(data)

	if err == nil {
		return nil
	}

	validatorErrs := err.(validator.ValidationErrors)

	for _, e := range validatorErrs {
		translatedErr := fmt.Errorf(
			e.Translate(v.trans),
		)
		strErrs = append(
			strErrs,
			translatedErr.Error(),
		)
	}

	errMsg := strings.Join(
		strErrs,
		", ",
	)

	validationErr := *entity.ErrValidation
	validationErr.Message = errMsg

	return &validationErr
}
