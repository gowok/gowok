package maps_test

import (
	"fmt"
	"testing"

	"github.com/golang-must/must"
	"github.com/gowok/gowok/maps"
	"github.com/ngamux/ngamux"
)

func ExampleToStruct() {
	type Config struct {
		Host string `json:"host"`
	}

	var m = map[string]any{
		"host": "localhost",
	}

	var c Config
	err := maps.ToStruct(m, &c)
	if err != nil {
		panic(err)
	}

	fmt.Println(c.Host)
	// Output:
	// localhost
}

func ExampleGet() {
	var m = map[string]any{
		"mantap": "jos",
		"gandos": map[string]any{
			"jos": "yoi",
		},
	}

	fmt.Println(maps.Get[string](m, "gandos.jos"))
	// Output:
	// yoi
}

func TestToStruct(t *testing.T) {
	type Config struct {
		Host    string `json:"host"`
		Enabled bool   `json:"enabled"`
	}
	var cases = []struct {
		Title string
		Input map[string]any
		Check func(*testing.T, error, Config)
	}{
		{
			"correct output to struct",
			map[string]any{"host": "localhost", "enabled": true},
			func(t *testing.T, err error, c Config) {
				must.Nil(t, err)
				must.Equal(t, c.Host, "localhost")
				must.True(t, c.Enabled)
			},
		},
		{
			"error if not possible",
			map[string]any{"host": "localhost", "enabled": func() bool { return true }},
			func(t *testing.T, err error, c Config) {
				must.NotNil(t, err)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Title, func(t *testing.T) {
			var output Config
			err := maps.ToStruct(c.Input, &output)
			c.Check(t, err, output)
		})
	}
}

func TestFromStruct(t *testing.T) {
	type Config struct {
		Host    string `json:"host"`
		Enabled any    `json:"enabled"`
	}
	var cases = []struct {
		Title string
		Input Config
		Check func(*testing.T, error, map[string]any)
	}{
		{
			"correct output from struct to map",
			Config{"localhost", true},
			func(t *testing.T, err error, c map[string]any) {
				must.Nil(t, err)
				must.Equal(t, c["host"], "localhost")
				must.Equal(t, c["enabled"], true)
			},
		},
		{
			"error if not possible",
			Config{"localhost", func() bool { return true }},
			func(t *testing.T, err error, c map[string]any) {
				must.NotNil(t, err)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Title, func(t *testing.T) {
			output, err := maps.FromStruct(c.Input)
			c.Check(t, err, output)
		})
	}

}

func TestGet(t *testing.T) {
	var cases = []struct {
		Title string
		Input map[string]any
		Check func(*testing.T, map[string]any)
	}{
		{
			"flat",
			map[string]any{"host": "localhost", "enabled": true},
			func(t *testing.T, c map[string]any) {
				must.Equal(t, maps.Get(c, "host", ""), "localhost")
				must.Equal(t, maps.Get(c, "enabled", false), true)
			},
		},
		{
			"nested",
			map[string]any{"host": "localhost", "address": map[string]any{"city": "Jakarta"}},
			func(t *testing.T, c map[string]any) {
				must.Equal(t, maps.Get(c, "host", ""), "localhost")
				must.Equal(t, maps.Get(c, "address.city", ""), "Jakarta")
				must.Equal(t, maps.Get(c, "host.name", ""), "")
			},
		},
		{
			"nested type alias",
			map[string]any{"host": "localhost", "address": ngamux.Map{"city": "Jakarta"}},
			func(t *testing.T, c map[string]any) {
				must.Equal(t, maps.Get(c, "host", ""), "localhost")
				must.Equal(t, maps.Get(c, "address.city", ""), "Jakarta")
				must.Equal(t, maps.Get(c, "host.name", ""), "")
			},
		},
		{
			"nested type any",
			map[string]any{"host": "localhost", "address": any(map[string]any{"city": "Jakarta"})},
			func(t *testing.T, c map[string]any) {
				must.Equal(t, maps.Get(c, "host", ""), "localhost")
				must.Equal(t, maps.Get(c, "address.city", ""), "Jakarta")
				must.Equal(t, maps.Get(c, "host.name", ""), "")
			},
		},
		{
			"default if not possible",
			map[string]any{"host": "localhost", "enabled": func() bool { return true }},
			func(t *testing.T, c map[string]any) {
				must.Equal(t, maps.Get(c, "host", ""), "localhost")
				must.Equal(t, maps.Get(c, "enabled", false), false)
				must.Equal(t, maps.Get(c, "uwu", 0), 0)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Title, func(t *testing.T) {
			c.Check(t, c.Input)
		})
	}
}
