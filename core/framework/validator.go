package framework

import (
	"context"
	"fmt"
	"github.com/SongOf/edge-storage-core/core/eserrors"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

type FilterValidateFunc func(value interface{}) bool
type FilterValueValidateFuncMap map[string]FilterValidateFunc

func NewValidator() (*Validator, error) {
	mValidator := validator.New()

	validatorObj := &Validator{
		validate:          mValidator,
		filterValidateMap: make(map[string]FilterValueValidateFuncMap),
	}

	err := mValidator.RegisterValidation("filter_key_validator", func(fl validator.FieldLevel) bool {
		tagValue := fl.Param()
		funcMap, ok := validatorObj.filterValidateMap[tagValue]

		filterName, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}
		_, ok = funcMap[filterName]
		if !ok {
			return false
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("register validation `filter_key_validator`:%v", err)
	}

	err = mValidator.RegisterValidation("filter_value_validator", func(fl validator.FieldLevel) bool {
		tagValue := fl.Param()
		funcMap, ok := validatorObj.filterValidateMap[tagValue]

		filterName := fl.Parent().String()
		validateFunc, ok := funcMap[filterName]
		if !ok {
			return false
		}

		return validateFunc(fl.Field().Interface())
	})
	if err != nil {
		return nil, fmt.Errorf("register validation `filter_value_validator`:%v", err)
	}

	mValidator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}
		return name
	})

	return validatorObj, nil
}

type Validator struct {
	validate          *validator.Validate
	filterValidateMap map[string]FilterValueValidateFuncMap
}

func (v *Validator) RegisterFilterValidator(set string, key string, validateFunc FilterValidateFunc) {
	funcMap, ok := v.filterValidateMap[set]
	if !ok {
		funcMap = make(map[string]FilterValidateFunc)
		v.filterValidateMap[set] = funcMap
	}

	funcMap[key] = validateFunc
}

//RegisterCustomValidatorTag register custom validator Tag and corresponding function
func (v *Validator) RegisterCustomValidatorTag(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn, false)
}

//ValidateParameters validate description
func (v *Validator) ValidateParameters(ctx context.Context, description ControllerDescription) error {
	if description == nil {
		return fmt.Errorf("description is nil")
	}

	validate := v.validate
	err := validate.Struct(description)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return fmt.Errorf("validate error:%v", err)
		}

		for _, err := range err.(validator.ValidationErrors) {
			return v.translate(err)
		}
	}

	return nil
}

func (v *Validator) translate(err validator.FieldError) error {
	fields := strings.SplitN(err.Namespace(), ".", 2)
	if len(fields) < 1 {
		// should not go here
		return fmt.Errorf("namespace is empty")
	}

	namespace := fields[len(fields)-1]

	var transErr error
	switch err.Tag() {
	case "required":
		transErr = eserrors.MissingParameter(namespace)

	case "oneof":
		transErr = eserrors.InvalidParameterValueEx(
			eserrors.InvalidParameterValueRangeCode,
			map[string]interface{}{
				"value":     err.Value(),
				"parameter": namespace,
				"range":     err.Param(),
			},
		)

	case "min":
		transErr = eserrors.InvalidParameterValueEx(
			eserrors.InvalidParameterValueTooSmallCode,
			map[string]interface{}{
				"value":     err.Value(),
				"parameter": namespace,
				"min":       err.Param(),
			},
		)

	case "max":
		transErr = eserrors.InvalidParameterValueEx(
			eserrors.InvalidParameterValueTooLargeCode,
			map[string]interface{}{
				"value":     err.Value(),
				"parameter": namespace,
				"max":       err.Param(),
			},
		)

	case "len":
		transErr = eserrors.InvalidParameterValueEx(
			eserrors.InvalidParameterValueLengthCode,
			map[string]interface{}{
				"value":     err.Value(),
				"parameter": namespace,
				"length":    err.Param(),
			},
		)

	case "filter_key_validator":
		transErr = eserrors.InvalidParameterValueEx(
			eserrors.InvalidParameterValueInvalidFilterCode,
			map[string]interface{}{
				"value": err.Value(),
			},
		)

	case "filter_value_validator":
		transErr = eserrors.InvalidParameterValueEx(
			eserrors.InvalidParameterValueInvalidFilterValueCode,
			map[string]interface{}{
				"value":     err.Value(),
				"parameter": namespace,
			},
		)

	case "eqfield":
		transErr = eserrors.InvalidParameterValueEx(
			eserrors.InvalidParameterValueFieldsCompareCode,
			map[string]interface{}{
				"leftField":  namespace,
				"rightField": err.Param(),
				"relation":   "equal to",
			},
		)

	case "nefield":
		transErr = eserrors.InvalidParameterValueEx(
			eserrors.InvalidParameterValueFieldsCompareCode,
			map[string]interface{}{
				"leftField":  namespace,
				"rightField": err.Param(),
				"relation":   "not equal to",
			},
		)

	case "gtfield":
		transErr = eserrors.InvalidParameterValueEx(
			eserrors.InvalidParameterValueFieldsCompareCode,
			map[string]interface{}{
				"leftField":  namespace,
				"rightField": err.Param(),
				"relation":   "greater than",
			},
		)

	case "gtefield":
		transErr = eserrors.InvalidParameterValueEx(
			eserrors.InvalidParameterValueFieldsCompareCode,
			map[string]interface{}{
				"leftField":  namespace,
				"rightField": err.Param(),
				"relation":   "greater than or equal to",
			},
		)

	case "ltfield":
		transErr = eserrors.InvalidParameterValueEx(
			eserrors.InvalidParameterValueFieldsCompareCode,
			map[string]interface{}{
				"leftField":  namespace,
				"rightField": err.Param(),
				"relation":   "less than",
			},
		)

	case "ltefield":
		transErr = eserrors.InvalidParameterValueEx(
			eserrors.InvalidParameterValueFieldsCompareCode,
			map[string]interface{}{
				"leftField":  namespace,
				"rightField": err.Param(),
				"relation":   "less than or equal to",
			},
		)

	default:
		transErr = eserrors.InvalidParameterValue(namespace, err.Value())
	}

	return transErr
}
