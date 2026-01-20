package json

import "encoding/json"

type marshaler func(any) ([]byte, error)
type unmarshaler func([]byte, any) error

var globalMarshaler = json.Marshal
var globalUnmarshaler = json.Unmarshal

func Marshal(v any) ([]byte, error) {
	return globalMarshaler(v)
}

func Unmarshal(data []byte, v any) error {
	return globalUnmarshaler(data, v)
}
