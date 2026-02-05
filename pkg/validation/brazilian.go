package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/marcelofabianov/wisp"
)

func RegisterBrazilianValidators(v Validator) error {
	validators := map[string]validator.Func{
		"cpf":   validateCPF,
		"cnpj":  validateCNPJ,
		"cep":   validateCEP,
		"phone": validatePhone,
		"email": validateEmail,
	}

	for tag, fn := range validators {
		if err := v.RegisterCustom(tag, fn); err != nil {
			return err
		}
	}

	return nil
}

func validateCPF(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	_, err := wisp.NewCPF(value)
	return err == nil
}

func validateCNPJ(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	_, err := wisp.NewCNPJ(value)
	return err == nil
}

func validateCEP(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	_, err := wisp.NewCEP(value)
	return err == nil
}

func validatePhone(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	_, err := wisp.NewPhone(value)
	return err == nil
}

func validateEmail(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	_, err := wisp.NewEmail(value)
	return err == nil
}
