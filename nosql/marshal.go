package nosql

import (
	"encoding"
	"encoding/json"
	"errors"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/buffer"
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
	"reflect"
)

var (
	ErrNotJSONMarshaller       = errors.New("not implements json.Marshaler")
	ErrNotEasyJSONMarhaller    = errors.New("not implements easyjson.EasyJSONMarshaller")
	ErrNotEasyJSONUnmarshaller = errors.New("not implements easyjson.EasyJSONUnmarshaller")
	ErrNotBinaryMarshaller     = errors.New("not implements encoding.BinaryMarshaler")
	ErrNotBinaryUnmarshaller   = errors.New("not implements encoding.BinaryUnmarshaler")
)

var (
	jsonMarshallerType       = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	jsonUnmarshallerType     = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	easyJsonMarshallerType   = reflect.TypeOf((*easyjson.Marshaler)(nil)).Elem()
	easyJsonUnmarshallerType = reflect.TypeOf((*easyjson.Unmarshaler)(nil)).Elem()
	binaryMarshallerType     = reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()
	binaryUnmarshallerType   = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
)

type Marshaller interface {
	Marshal(unmarshalled interface{}, into []byte) ([]byte, error)

	Unmarshal(data []byte, unmarshalled interface{}) error
}

func MarshallerOfType(t reflect.Type) Marshaller {
	if t.Implements(easyJsonMarshallerType) && t.Implements(easyJsonUnmarshallerType) {
		return EasyJsonMarshaller{}
	}
	if t.Implements(binaryMarshallerType) && t.Implements(binaryUnmarshallerType) {
		return BinaryMarshaller{}
	}
	return JsonMarshaller{}
}

func JsonMarshallerOfType(t reflect.Type) Marshaller {
	if t.AssignableTo(easyJsonMarshallerType) && t.AssignableTo(easyJsonUnmarshallerType) {
		return EasyJsonMarshaller{}
	}
	return JsonMarshaller{}
}

func JsonMarshallerOf(unmarshalled interface{}) Marshaller {
	// Use EasyJSON if available
	if _, ok := unmarshalled.(easyjson.MarshalerUnmarshaler); ok {
		return EasyJsonMarshaller{}
	}
	return JsonMarshaller{}
}

func BinaryMarshallerOf(unmarshalled interface{}) Marshaller {
	if _, ok := unmarshalled.(encoding.BinaryUnmarshaler); !ok {
		return nil
	}
	if _, ok := unmarshalled.(BinaryMarshallerInto); ok {
		return BinaryNoAllocMarshaller{}
	}
	if _, ok := unmarshalled.(encoding.BinaryMarshaler); !ok {
		return nil
	}
	return BinaryMarshaller{}
}

func MarshallerOf(unmarshalled interface{}) Marshaller {
	switch unmarshalled.(type) {
	case json.Marshaler:
		return JsonMarshaller{}
	case easyjson.Marshaler:
		return EasyJsonMarshaller{}
	case encoding.BinaryMarshaler:
		return &BinaryMarshaller{}
	default:
		return JsonMarshaller{}
	}
}

type JsonMarshaller struct {
}

func (s JsonMarshaller) Marshal(unmarshalled interface{}, into []byte) ([]byte, error) {
	switch t := unmarshalled.(type) {
	case json.Marshaler:
		var (
			data []byte
			err  error
		)
		if data, err = t.MarshalJSON(); err != nil {
			return nil, err
		}
		return data, nil
	default:
		var (
			data []byte
			err  error
		)
		if data, err = json.Marshal(unmarshalled); err != nil {
			return nil, err
		}
		return data, nil
	}
}

func (s JsonMarshaller) Unmarshal(data []byte, unmarshalled interface{}) error {
	switch t := unmarshalled.(type) {
	case json.Unmarshaler:
		return t.UnmarshalJSON(data)
	case easyjson.Unmarshaler:
		lexer := jlexer.Lexer{
			Data:              data,
			UseMultipleErrors: false,
		}
		t.UnmarshalEasyJSON(&lexer)
		return lexer.Error()
	default:
		return json.Unmarshal(data, unmarshalled)
	}
}

type EasyJsonMarshaller struct {
}

func (s EasyJsonMarshaller) Marshal(unmarshalled interface{}, into []byte) ([]byte, error) {
	switch t := unmarshalled.(type) {
	case easyjson.Marshaler:
		writer := jwriter.Writer{
			Buffer: buffer.Buffer{
				Buf: into,
			},
		}
		t.MarshalEasyJSON(&writer)
		return writer.BuildBytes()
	default:
		return into, ErrNotEasyJSONMarhaller
	}
}

func (s EasyJsonMarshaller) Unmarshal(data []byte, unmarshalled interface{}) error {
	switch t := unmarshalled.(type) {
	case easyjson.Unmarshaler:
		lexer := jlexer.Lexer{
			Data:              data,
			UseMultipleErrors: false,
		}
		t.UnmarshalEasyJSON(&lexer)
		return lexer.Error()
	default:
		return ErrNotEasyJSONUnmarshaller
	}
}

type BinaryMarshaller struct {
}

func (s BinaryMarshaller) Marshal(unmarshalled interface{}, into []byte) ([]byte, error) {
	switch t := unmarshalled.(type) {
	case encoding.BinaryMarshaler:
		return t.MarshalBinary()
	default:
		return into, ErrNotBinaryMarshaller
	}
}

func (s BinaryMarshaller) Unmarshal(data []byte, unmarshalled interface{}) error {
	switch t := unmarshalled.(type) {
	case encoding.BinaryUnmarshaler:
		return t.UnmarshalBinary(data)
	default:
		return ErrNotBinaryUnmarshaller
	}
}

type BinaryNoAllocMarshaller struct {
}

type BinaryMarshallerInto interface {
	MarshalBinaryInto(into []byte) ([]byte, error)
}

func (s BinaryNoAllocMarshaller) Marshal(unmarshalled interface{}, into []byte) ([]byte, error) {
	switch t := unmarshalled.(type) {
	case encoding.BinaryMarshaler:
		return t.MarshalBinary()
	default:
		return into, ErrNotBinaryMarshaller
	}
}

func (s BinaryNoAllocMarshaller) Unmarshal(data []byte, unmarshalled interface{}) error {
	switch t := unmarshalled.(type) {
	case encoding.BinaryUnmarshaler:
		return t.UnmarshalBinary(data)
	default:
		return ErrNotBinaryUnmarshaller
	}
}
