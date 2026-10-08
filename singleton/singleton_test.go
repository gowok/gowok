package singleton

import (
	"sync"
	"testing"

	"github.com/golang-must/must"
)

func TestSingleton(t *testing.T) {
	testCases := []struct {
		name   string
		action func(t *testing.T)
	}{
		{
			name: "positive/lazy initialization",
			action: func(t *testing.T) {
				expected := 10
				calls := 0
				s := New(func() int {
					calls++
					return expected
				})

				must.Equal(t, 0, calls)
				val1 := s()
				must.Equal(t, 1, calls)
				must.Equal(t, expected, *val1)

				val2 := s()
				must.Equal(t, 1, calls)
				must.Equal(t, val1, val2)
			},
		},
		{
			name: "positive/override value",
			action: func(t *testing.T) {
				s := New(func() int { return 10 })

				val1 := s()
				must.Equal(t, 10, *val1)

				override := 20
				val2 := s(override)
				must.Equal(t, override, *val2)

				val3 := s()
				must.Equal(t, override, *val3)
			},
		},
		{
			name: "positive/concurrency",
			action: func(t *testing.T) {
				calls := 0
				s := New(func() int {
					calls++
					return 100
				})

				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						s()
					}()
				}
				wg.Wait()

				must.Equal(t, 1, calls)
			},
		},
		{
			name: "negative/multiple override values",
			action: func(t *testing.T) {
				s := New(func() int { return 10 })

				val := s(20, 30)
				must.Equal(t, 20, *val)
			},
		},
		{
			name: "positive/struct type",
			action: func(t *testing.T) {
				type User struct {
					Name string
				}
				s := New(func() User {
					return User{Name: "Alice"}
				})

				val1 := s()
				must.Equal(t, "Alice", val1.Name)

				s(User{Name: "Bob"})
				val2 := s()
				must.Equal(t, "Bob", val2.Name)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.action(t)
		})
	}
}
