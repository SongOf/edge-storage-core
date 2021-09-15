package eserrors

const (
	InvalidParameterSyntaxCode        = "Syntax"
	InvalidParameterSyntaxCodeMessage = "The json syntax of your input is invalid."

	InvalidParameterValueTypeCode    = "Type"
	InvalidParameterValueTypeMessage = "The type of parameter `{{.parameter}}` is `{{.actualType}}`, " +
		"expect `{{.expectType}}`."

	InvalidParameterBodyTooLargeCode    = "BodyTooLarge"
	InvalidParameterBodyTooLargeMessage = "The request body is too large."

	InvalidParameterValueRangeCode    = "Range"
	InvalidParameterValueRangeMessage = "The value `{{.value}}` specified in the parameter `{{.parameter}}` " +
		"must be range in [{{.range}}]."

	InvalidParameterValueTooSmallCode    = "TooSmall"
	InvalidParameterValueTooSmallMessage = "The value `{{.value}}` specified in the parameter `{{.parameter}}` " +
		"must bigger than `{{.min}}`."

	InvalidParameterValueTooLargeCode    = "TooLarge"
	InvalidParameterValueTooLargeMessage = "The value `{{.value}}` specified in the parameter `{{.parameter}}` " +
		"must less than `{{.max}}`."

	InvalidParameterValueLengthCode    = "Length"
	InvalidParameterValueLengthMessage = "The value `{{.value}}` specified in the parameter `{{.parameter}}` " +
		"which length must be `{{.length}}`."

	InvalidParameterValueInvalidFilterCode    = "InvalidFilter"
	InvalidParameterValueInvalidFilterMessage = "The specified filter `{{.value}}` is not valid."

	InvalidParameterValueInvalidFilterValueCode    = "InvalidFilterValues"
	InvalidParameterValueInvalidFilterValueMessage = "The filter value `{{.value}}` specified in the " +
		"parameter {{.parameter}} is not valid."

	InvalidParameterValueFieldsCompareCode    = "FieldsCompare"
	InvalidParameterValueFieldsCompareMessage = "`{{.leftField}}` must {{.relation}} `{{.rightField}}`"

	Aaa = "Aaa"
	Bbb = "Bbb"
)
