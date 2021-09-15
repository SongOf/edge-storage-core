package eserrors

func InternalError() EsError {
	const (
		DefaultErrorCode    = "InternalError"
		DefaultErrorMessage = "An internal error has occurred. Retry your request"
	)
	return &baseError{
		Code:            DefaultErrorCode,
		Message:         DefaultErrorMessage,
		MessageTemplate: "",
		SecondaryCode:   "",
		Data:            nil,
	}
}

func InvalidParameterEx(secondaryCode string, data interface{}) EsError {
	const InvalidParameterCode = "InvalidParameter"

	return &baseError{
		Code:            InvalidParameterCode,
		Message:         errorMap[secondaryCode],
		MessageTemplate: errorMap[secondaryCode],
		SecondaryCode:   secondaryCode,
		Data:            data,
	}
}

func InvalidParameterValueEx(secondaryCode string, data interface{}) EsError {
	const InvalidParameterValueCode = "InvalidParameterValue"

	return &baseError{
		Code:            InvalidParameterValueCode,
		Message:         errorMap[secondaryCode],
		MessageTemplate: errorMap[secondaryCode],
		SecondaryCode:   secondaryCode,
		Data:            data,
	}
}

func UnknownParameter(parameterName string) EsError {
	const (
		UnknownParameterCode    = "UnknownParameter"
		UnknownParameterMessage = "The parameter `{{.ParameterName}}` is not recognized."
	)
	return &baseError{
		Code:            UnknownParameterCode,
		Message:         UnknownParameterMessage,
		MessageTemplate: UnknownParameterMessage,
		SecondaryCode:   "",
		Data:            struct{ ParameterName string }{ParameterName: parameterName},
	}
}

func MissingParameter(parameter string) EsError {
	const (
		MissingParameterCode    = "MissingParameter"
		MissingParameterMessage = "The request is missing a required parameter `{{.Parameter}}`."
	)
	return &baseError{
		Code:            MissingParameterCode,
		Message:         MissingParameterMessage,
		MessageTemplate: MissingParameterMessage,
		SecondaryCode:   "",
		Data:            struct{ Parameter string }{Parameter: parameter},
	}
}

func InvalidAction(action string) EsError {
	const (
		InvalidActionCode    = "InvalidAction"
		InvalidActionMessage = "The action `{{.Action}}` requested is not found."
	)
	return &baseError{
		Code:            InvalidActionCode,
		Message:         InvalidActionMessage,
		MessageTemplate: InvalidActionMessage,
		SecondaryCode:   "",
		Data:            struct{ Action string }{Action: action},
	}
}

func InvalidParameterValue(parameter string, value interface{}) EsError {
	const (
		InvalidParameterValueCode    = "InvalidParameterValue"
		InvalidParameterValueMessage = "The value `{{.Value}}` specified in the parameter `{{.Parameter}}` is not valid."
	)
	return &baseError{
		Code:            InvalidParameterValueCode,
		Message:         InvalidParameterValueMessage,
		MessageTemplate: InvalidParameterValueMessage,
		SecondaryCode:   "",
		Data: struct {
			Parameter string
			Value     interface{}
		}{Parameter: parameter, Value: value},
	}
}
