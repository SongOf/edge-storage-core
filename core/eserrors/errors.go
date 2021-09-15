package eserrors

import (
	"bytes"
	"fmt"
	"text/template"
)

type EsError interface {
	error
	Format() (code string, message string)
	Wrap(error) EsError
	Unwrap() error
	GetData() interface{}
}

var errorMap = map[string]string{
	Aaa: Bbb,
}

func InitErrorMap(input map[string]string) {
	for k, v := range input {
		if _, ok := errorMap[k]; ok {
			// don't panic
			fmt.Printf("input error map item override framework, error code: [%s]", k)
		}
		errorMap[k] = v
	}
	return
}

type baseError struct {
	Code            string
	Message         string
	MessageTemplate string      `json:"-"`
	SecondaryCode   string      `json:"-"`
	Data            interface{} `json:"-"`
	err             error
}

func (be *baseError) Format() (string, string) {
	code, msg := be.Code, be.Message

	if be.SecondaryCode != "" {
		code = be.Code + "." + be.SecondaryCode
	}

	if be.MessageTemplate != "" && be.Data != nil {
		formatMessage := format(be.MessageTemplate, be.Data)
		if formatMessage != "" {
			msg = formatMessage
		}
	}

	return code, msg
}

func format(messageTemplate string, meta interface{}) string {
	t := template.Must(template.New("messageTemplate").Parse(messageTemplate))
	var buffer bytes.Buffer
	if err := t.Execute(&buffer, meta); err == nil {
		message := buffer.String()
		return message
	}
	return ""
}

func (be *baseError) Error() string {
	code, msg := be.Format()
	return code + ":" + msg
}

func (be *baseError) Unwrap() error {
	return be.err
}

func (be *baseError) Wrap(err error) EsError {
	be.err = err
	return be
}

func (be *baseError) GetData() interface{} {
	return be.Data
}
