package sql

import (
	"encoding/json"
	"testing"

	"github.com/golang-must/must"
)

func TestNewNull(t *testing.T) {
	t.Run("positive/int", func(t *testing.T) {
		val := 10
		n := NewNull(&val)
		must.True(t, n.Valid)
		must.Equal(t, val, n.V)
	})

	t.Run("negative/nil input", func(t *testing.T) {
		var val *int
		n := NewNull(val)
		must.False(t, n.Valid)
	})
}

func TestNull_MarshalJSON(t *testing.T) {
	testCases := []struct {
		name     string
		input    Null[int]
		expected string
	}{
		{
			name: "positive/valid int",
			input: func() Null[int] {
				v := 42
				return NewNull(&v)
			}(),
			expected: "42",
		},
		{
			name:     "positive/nil input",
			input:    NewNull[int](nil),
			expected: "null",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.input)
			must.Nil(t, err)
			must.Equal(t, tc.expected, string(b))
		})
	}
}

func TestNull_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		wantVal   int
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "positive/valid json",
			input:     "123",
			wantVal:   123,
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "positive/null json",
			input:     "null",
			wantVal:   0,
			wantValid: false,
			wantErr:   false,
		},
		{
			name:    "negative/invalid type",
			input:   `"not_an_int"`,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var n Null[int]
			err := json.Unmarshal([]byte(tc.input), &n)
			if tc.wantErr {
				must.NotNil(t, err)
			} else {
				must.Nil(t, err)
				must.Equal(t, tc.wantValid, n.Valid)
				if n.Valid {
					must.Equal(t, tc.wantVal, n.V)
				}
			}
		})
	}
}

func TestNull_StringMarshal(t *testing.T) {
	t.Run("positive/string", func(t *testing.T) {
		val := "test"
		n := NewNull(&val)
		b, err := json.Marshal(n)
		must.Nil(t, err)
		must.Equal(t, `"test"`, string(b))

		var n2 Null[string]
		err = json.Unmarshal(b, &n2)
		must.Nil(t, err)
		must.True(t, n2.Valid)
		must.Equal(t, val, n2.V)
	})
}
