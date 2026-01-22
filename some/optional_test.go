package some

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang-must/must"
)

func TestEmpty(t *testing.T) {
	s := Empty[string]()
	must.False(t, s.IsPresent())
	val, ok := s.Get()
	must.False(t, ok)
	must.Equal(t, "", val)
}

func TestOf(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		isPresent bool
	}{
		{"positive/string", "hello", true},
		{"positive/int", 123, true},
		{"negative/nil func", (func())(nil), false},
		{"positive/valid func", func() {}, true},
		{"negative/nil pointer", (*int)(nil), false},
		{"positive/valid pointer", new(int), true},
		{"negative/nil interface", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Of(tt.input)
			must.Equal(t, tt.isPresent, s.IsPresent())
			if tt.isPresent {
				val, ok := s.Get()
				must.True(t, ok)
				must.Equal(t, tt.input, val)
			}
		})
	}
}

func TestSome_IsPresent(t *testing.T) {
	must.True(t, Of("test").IsPresent())
	must.False(t, Empty[string]().IsPresent())
}

func TestSome_OrElse(t *testing.T) {
	must.Equal(t, "test", Of("test").OrElse("default"))
	must.Equal(t, "default", Empty[string]().OrElse("default"))
}

func TestSome_OrElseFunc(t *testing.T) {
	must.Equal(t, "test", Of("test").OrElseFunc(func() string { return "default" }))
	must.Equal(t, "default", Empty[string]().OrElseFunc(func() string { return "default" }))
}

func TestSome_Get(t *testing.T) {
	tests := []struct {
		name    string
		s       Some[string]
		wantVal string
		wantOk  bool
	}{
		{"positive/present", Of("test"), "test", true},
		{"negative/not present", Empty[string](), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := tt.s.Get()
			must.Equal(t, tt.wantOk, ok)
			must.Equal(t, tt.wantVal, val)
		})
	}
}

func TestSome_IfPresent(t *testing.T) {
	tests := []struct {
		name            string
		s               Some[string]
		hasElseCallback bool
		wantCalled      bool
		wantElseCalled  bool
	}{
		{"positive/present with callback", Of("test"), true, true, false},
		{"negative/not present with else callback", Empty[string](), true, false, true},
		{"negative/not present without else callback", Empty[string](), false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			elseCalled := false

			callback := func(v string) {
				called = true
				must.Equal(t, "test", v)
			}

			if tt.hasElseCallback {
				tt.s.IfPresent(callback, func() {
					elseCalled = true
				})
			} else {
				tt.s.IfPresent(callback)
			}

			must.Equal(t, tt.wantCalled, called)
			must.Equal(t, tt.wantElseCalled, elseCalled)
		})
	}
}

func TestSome_OrPanic(t *testing.T) {
	tests := []struct {
		name      string
		s         Some[string]
		errs      []error
		wantVal   string
		wantPanic bool
		panicErr  string
	}{
		{
			name:    "positive/present",
			s:       Of("test"),
			wantVal: "test",
		},
		{
			name:      "negative/not present default panic",
			s:         Empty[string](),
			wantPanic: true,
		},
		{
			name:      "negative/not present custom panic",
			s:         Empty[string](),
			errs:      []error{errors.New("custom error")},
			wantPanic: true,
			panicErr:  "custom error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					r := recover()
					must.NotNil(t, r)
					if tt.panicErr != "" {
						must.Equal(t, tt.panicErr, r.(error).Error())
					}
				}()
			}

			val := tt.s.OrPanic(tt.errs...)
			if !tt.wantPanic {
				must.Equal(t, tt.wantVal, val)
			}
		})
	}
}

func TestSome_JSON(t *testing.T) {
	type container struct {
		Opt Some[string] `json:"opt"`
	}

	tests := []struct {
		name      string
		operation string
		input     any
		want      string
		wantErr   bool
	}{
		{
			name:      "positive/marshal present",
			operation: "marshal",
			input:     Of("test"),
			want:      "\"test\"",
		},
		{
			name:      "positive/marshal empty",
			operation: "marshal",
			input:     Empty[string](),
			want:      "{}",
		},
		{
			name:      "positive/marshal embedded",
			operation: "marshal",
			input:     container{Opt: Of("test")},
			want:      "{\"opt\":\"test\"}",
		},
		{
			name:      "positive/unmarshal present",
			operation: "unmarshal",
			input:     "\"test\"",
			want:      "test",
		},
		{
			name:      "positive/unmarshal empty object",
			operation: "unmarshal",
			input:     "{}",
		},
		{
			name:      "positive/unmarshal null",
			operation: "unmarshal",
			input:     "null",
		},
		{
			name:      "negative/unmarshal error",
			operation: "unmarshal",
			input:     "{invalid}",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.operation == "marshal" {
				b, err := json.Marshal(tt.input)
				must.Nil(t, err)
				must.Equal(t, tt.want, string(b))
			} else {
				var s Some[string]
				err := json.Unmarshal([]byte(tt.input.(string)), &s)
				if tt.wantErr {
					must.NotNil(t, err)
				} else {
					must.Nil(t, err)
					if tt.want != "" {
						must.True(t, s.IsPresent())
						val, _ := s.Get()
						must.Equal(t, tt.want, val)
					} else {
						must.False(t, s.IsPresent())
					}
				}
			}
		})
	}
}
