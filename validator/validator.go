package validator

import (
	"reflect"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/vi"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var trans ut.Translator

var validate *validator.Validate

type ValidatorConfig struct {
	*validator.Validate
}

type ValidatorClientConfig struct {
	dicionario map[string]map[string]string
	traducao   map[string]map[string]string
}

type ValidatorOption func(d *ValidatorClientConfig)

func WithDicionario(dicionario map[string]map[string]string) ValidatorOption {
	return func(c *ValidatorClientConfig) {
		c.dicionario = dicionario
	}
}

func WithTraducoes(traducao map[string]map[string]string) ValidatorOption {
	return func(c *ValidatorClientConfig) {
		c.traducao = traducao
	}
}

func translationFunc(t ut.Translator, fe validator.FieldError) string {
	field, err := t.T(fe.Field())
	if err != nil {
		field = fe.Field()
	}
	msg, err := t.T(fe.Tag(), field, fe.Param())
	if err != nil {
		return fe.Error()
	}
	return msg
}

func Initialize(opts ...ValidatorOption) {
	validatorOptions := &ValidatorClientConfig{}
	for _, opt := range opts {
		opt(validatorOptions)
	}

	v := validator.New()
	enLocale := en.New()
	utrans := ut.New(enLocale, enLocale, vi.New())
	trans, _ = utrans.GetTranslator("pt")

	v.RegisterValidation("ISO8601date", isISO8601Date)
	v.RegisterValidation("PhoneFormat", phoneFormat)
	v.RegisterValidation("CPForCNPJ", validateCPForCNPJ)

	for locale, dict := range validatorOptions.dicionario {
		engine, _ := utrans.FindTranslator(locale)
		for key, trans := range dict {
			_ = engine.Add(key, trans, false)
		}
	}

	for locale, translation := range validatorOptions.traducao {
		engine, _ := utrans.FindTranslator(locale)
		for tag, trans := range translation {
			_ = v.RegisterTranslation(tag, engine, func(t ut.Translator) error {
				return t.Add(tag, trans, false)
			}, translationFunc)
		}
	}

	validate = v
}

func ValidateStruct(payload any) error {
	if err := validate.Struct(payload); err != nil {
		return handleValidatorFieldError(payload, err)
	}

	return nil
}

func HandleValidatorFieldError(data any, err error) error {
	return handleValidatorFieldError(data, err)
}

func handleValidatorFieldError(data any, err error) error {
	errs := err.(validator.ValidationErrors)

	requestError := RequestError{}

	for _, e := range errs {
		field := Field{
			Field:   getFormatedField(data, e.Field()),
			Message: e.Translate(trans),
			Errs:    e.Error(),
		}

		requestError.Fields = append(requestError.Fields, field)
	}

	return requestError
}

func getFormatedField(data any, field string) string {
	if field, ok := reflect.TypeOf(data).Elem().FieldByName(field); ok {
		return field.Tag.Get("json")
	}

	return ""
}
