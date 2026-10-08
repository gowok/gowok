package gowok

import (
	"testing"

	"github.com/golang-must/must"
)

func TestToPtr(t *testing.T) {
	t.Run("positive/int", func(t *testing.T) {
		val := 10
		ptr := ToPtr(val)
		must.NotNil(t, ptr)
		must.Equal(t, val, *ptr)
	})

	t.Run("positive/string", func(t *testing.T) {
		val := "hello"
		ptr := ToPtr(val)
		must.NotNil(t, ptr)
		must.Equal(t, val, *ptr)
	})
}

func TestDD(t *testing.T) {
	t.Run("positive/simple map", func(t *testing.T) {
		input := map[string]string{"foo": "bar"}
		err := DD(input)
		must.Nil(t, err)
	})
}
