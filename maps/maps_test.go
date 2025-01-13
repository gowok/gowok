package maps_test

import (
	"fmt"

	"github.com/gowok/gowok/maps"
)

func ExampleMapToStruct() {
	type Config struct {
		Host string `json:"host"`
	}

	var m = map[string]any{
		"host": "localhost",
	}

	var c Config
	maps.MapToStruct(m, &c)

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
