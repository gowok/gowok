package json

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang-must/must"
)

func TestMarshal(t *testing.T) {
	testCases := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name:    "positive/basic struct",
			input:   map[string]string{"foo": "bar"},
			wantErr: false,
		},
		{
			name:    "negative/unmarshalable",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.input)
			if tc.wantErr {
				must.NotNil(t, err)
			} else {
				must.Nil(t, err)
				must.NotNil(t, b)
				var m map[string]any
				must.Nil(t, json.Unmarshal(b, &m))
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "positive/valid json",
			input:   `{"foo":"bar"}`,
			wantErr: false,
		},
		{
			name:    "negative/invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var m map[string]any
			err := Unmarshal([]byte(tc.input), &m)
			if tc.wantErr {
				must.NotNil(t, err)
			} else {
				must.Nil(t, err)
				must.Equal(t, "bar", m["foo"])
			}
		})
	}
}

func TestConfigure(t *testing.T) {
	oldMarshaler := globalMarshaler
	oldUnmarshaler := globalUnmarshaler
	defer func() {
		globalMarshaler = oldMarshaler
		globalUnmarshaler = oldUnmarshaler
	}()

	t.Run("positive/custom marshaler", func(t *testing.T) {
		customErr := errors.New("custom marshal error")
		Configure(func(v any) ([]byte, error) {
			return nil, customErr
		}, nil)

		_, err := Marshal("anything")
		must.Equal(t, customErr, err)
	})

	t.Run("positive/custom unmarshaler", func(t *testing.T) {
		customErr := errors.New("custom unmarshal error")
		Configure(nil, func(data []byte, v any) error {
			return customErr
		})

		err := Unmarshal([]byte("{}"), &struct{}{})
		must.Equal(t, customErr, err)
	})
}
