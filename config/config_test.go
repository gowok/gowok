package config

import (
	"testing"

	"github.com/golang-must/must"
)

func TestConfig_Map(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func()
		teardown func()
		check    func(t *testing.T, m map[string]any)
	}{
		{
			name: "positive/empty map",
			setup: func() {
				for k := range ConfigMap {
					delete(ConfigMap, k)
				}
			},
			check: func(t *testing.T, m map[string]any) {
				must.NotNil(t, m)
				must.Equal(t, 0, len(m))
			},
		},
		{
			name: "positive/map with values",
			setup: func() {
				ConfigMap["key1"] = "val1"
				ConfigMap["key2"] = 123
			},
			teardown: func() {
				delete(ConfigMap, "key1")
				delete(ConfigMap, "key2")
			},
			check: func(t *testing.T, m map[string]any) {
				must.Equal(t, "val1", m["key1"])
				must.Equal(t, 123, m["key2"])
			},
		},
	}

	c := &Config{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}
			if tc.teardown != nil {
				defer tc.teardown()
			}

			tc.check(t, c.Map())
		})
	}
}
