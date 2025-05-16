package singleton

import (
	"testing"

	"github.com/golang-must/must"
)

func TestSingleton(t *testing.T) {
	cases := []struct {
		Description string
		Expected    int
	}{
		{
			Description: "simple",
			Expected:    10,
		},
	}
	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			UUU := New(func() int {
				return c.Expected
			})
			must.Equal(t, UUU(), &c.Expected)
		})
	}
}
