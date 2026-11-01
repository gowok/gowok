package gowok

import (
	"errors"
	"testing"

	"github.com/golang-must/must"
)

func TestAsync_AllResults(t *testing.T) {
	t.Run("negative/keeps successful results", func(t *testing.T) {
		res, err := Async.AllResults(
			func() (any, error) { return 1, nil },
			func() (any, error) { return nil, errors.New("err") },
			func() (any, error) { return 3, nil },
		)
		must.NotNil(t, err)
		must.Equal(t, 3, len(res))
		must.Equal(t, 1, res[0].(int))
		must.Nil(t, res[1])
		must.Equal(t, 3, res[2].(int))
	})

	t.Run("positive/all succeed", func(t *testing.T) {
		res, err := Async.AllResults(
			func() (any, error) { return 1, nil },
			func() (any, error) { return 2, nil },
		)
		must.Nil(t, err)
		must.Equal(t, 2, len(res))
		must.Equal(t, 1, res[0].(int))
		must.Equal(t, 2, res[1].(int))
	})
}

func TestAsync_All(t *testing.T) {
	testCases := []struct {
		name      string
		tasks     []func() (any, error)
		wantCount int
		wantErr   bool
	}{
		{
			name: "positive/success",
			tasks: []func() (any, error){
				func() (any, error) { return 1, nil },
				func() (any, error) { return 2, nil },
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "negative/error",
			tasks: []func() (any, error){
				func() (any, error) { return 1, nil },
				func() (any, error) { return nil, errors.New("err") },
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Async.All(tc.tasks...)
			if tc.wantErr {
				must.NotNil(t, err)
				must.Nil(t, res)
			} else {
				must.Nil(t, err)
				must.Equal(t, tc.wantCount, len(res))
			}
		})
	}
}
