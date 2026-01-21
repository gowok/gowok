package gowok

import (
	"errors"
	"testing"

	"github.com/golang-must/must"
)

func TestMust(t *testing.T) {
	t.Run("positive/no error", func(t *testing.T) {
		val := "success"
		res := Must(val, nil)
		must.Equal(t, val, res)
	})

	t.Run("negative/with error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()

		Must("failure", errors.New("something went wrong"))
	})
}
