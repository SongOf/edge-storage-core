package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SongOf/edge-storage-core/core/eserrors"
	"github.com/mitchellh/mapstructure"
	"regexp"
	"strings"
)

type Parser struct {
	// regExpObj *regexp.Regexp
	decodeFuncErrorRegexp  *regexp.Regexp
	decodeSliceErrorRegexp *regexp.Regexp
	decodeMapErrorRegexp   *regexp.Regexp
}

//NewParser Create a parser TestObject
func NewParser() (*Parser, error) {
	compiledDecodeFuncErrorRegexp, err := regexp.Compile(
		`'([0-9A-Za-z_\[\].]+?)' expected type '(\w+?)', got unconvertible type '(\w+?)'`,
	)
	if err != nil {
		return nil, fmt.Errorf("create compiledDecodeFuncErrorRegexp:%v", err)
	}

	compiledDecodeSliceFuncErrorRegexp, err := regexp.Compile(
		`'([0-9A-Za-z_\[\].]+?)': source data must be an (\w+?) or slice, got (.+?)$`,
	)
	if err != nil {
		return nil, fmt.Errorf("create compiledDecodeSliceFuncErrorRegexp:%v", err)
	}

	compiledDecodeMapErrorRegexp, err := regexp.Compile(
		`'([0-9A-Za-z_\[\].]+?)' expected a (\w+?), got '(.+?)'$`,
	)

	return &Parser{
		decodeFuncErrorRegexp:  compiledDecodeFuncErrorRegexp,
		decodeSliceErrorRegexp: compiledDecodeSliceFuncErrorRegexp,
		decodeMapErrorRegexp:   compiledDecodeMapErrorRegexp,
	}, nil
}

//PreParseRequest decode raw json to a map
func (parser *Parser) PreParseRequest(rawJsonBody []byte) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if err := json.Unmarshal(rawJsonBody, &params); err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			return nil, eserrors.InvalidParameterEx(eserrors.InvalidParameterSyntaxCode, nil)
		default:
			return nil, fmt.Errorf("unmarshal json error:%v", err)
		}
	}

	return params, nil
}

//CheckParams check unused parameters, if exists, return an UnknownParameter
func (parser *Parser) CheckParams(ctx context.Context, params map[string]interface{},
	description ControllerDescription) error {

	if description == nil {
		return fmt.Errorf("description is nil")
	}
	var metadata mapstructure.Metadata
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Squash: true,
		// ErrorUnused: true,
		Metadata: &metadata,
		Result:   description,
	})

	if err != nil {
		return fmt.Errorf("create mapstructure decoder error:%v", err)
	}

	err = decoder.Decode(params)

	if err != nil {
		newInvalidParameterValueTypeError := func(parameter, actualType, expectType string) error {
			return eserrors.InvalidParameterValueEx(
				eserrors.InvalidParameterValueTypeCode,
				map[string]string{
					"parameter":  parameter,
					"actualType": actualType,
					"expectType": expectType,
				})
		}

		if decoderErr, ok := err.(*mapstructure.Error); !ok {
			return fmt.Errorf("decoder decode error:%v", err)
		} else {
			for _, wrappedError := range decoderErr.WrappedErrors() {
				if res := parser.decodeFuncErrorRegexp.FindStringSubmatch(wrappedError.Error()); len(res) == 4 {
					return newInvalidParameterValueTypeError(res[1], res[3], res[2])
				} else if res := parser.decodeSliceErrorRegexp.FindStringSubmatch(wrappedError.Error()); len(res) == 4 {
					return newInvalidParameterValueTypeError(res[1], res[3], res[2])
				} else if res := parser.decodeMapErrorRegexp.FindStringSubmatch(wrappedError.Error()); len(res) == 4 {
					return newInvalidParameterValueTypeError(res[1], res[3], res[2])
				}
			}
			return err
		}
	}

	// return UnknownParameter if params have unused fields
	if len(metadata.Unused) > 0 {
		return eserrors.UnknownParameter(metadata.Unused[0])
	}

	return nil
}

//ParseRequest decode raw json to a description struct
func (parser *Parser) ParseRequest(rawJsonReqParams []byte, description ControllerDescription) error {

	if description == nil {
		return fmt.Errorf("description is nil")
	}

	err := json.Unmarshal(rawJsonReqParams, description)
	if err != nil {
		switch jerr := err.(type) {
		case *json.UnmarshalTypeError:
			return parser.translateJsonUnmarshalTypeError(jerr)
		case *json.SyntaxError:
			return eserrors.InvalidParameterEx(eserrors.InvalidParameterSyntaxCode, nil)

		default:
			return fmt.Errorf("unmarshal json error:%v", err)
		}
	}

	return nil
}

func (parser *Parser) translateJsonUnmarshalTypeError(typeErr *json.UnmarshalTypeError) error {
	translateType := func(t string) string {
		hasPrefix := func(prefix string) bool {
			return strings.HasPrefix(t, prefix)
		}

		switch {
		case hasPrefix("string"):
			return "string"
		case hasPrefix("number"), hasPrefix("uint"), hasPrefix("int"), hasPrefix("float"):
			return "number"
		case hasPrefix("[]"):
			return "array"
		default:
			return "TestObject"
		}
	}

	return eserrors.InvalidParameterValueEx(
		eserrors.InvalidParameterValueTypeCode,
		map[string]interface{}{
			"parameter":  typeErr.Field,
			"actualType": typeErr.Value,
			"expectType": translateType(typeErr.Type.String()),
		},
	)
}
