package maps

import (
	"encoding/json"
	"strings"
)

// MapToStruct is a helper function to convert map[string]any to struct
func MapToStruct(data any, v any) error {
	jsoned, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsoned, v)
	if err != nil {
		return err
	}

	return nil
}

// Get is a helper function to get value from map[string]any
// it's posible to get value from nested map by using dot (.) as separator
func Get[T any](data map[string]any, path string, defaults ...T) T {
	var value T
	if len(defaults) > 0 {
		value = defaults[0]
	}

	keys := strings.Split(path, ".")
	current := data

	for i, key := range keys {
		val, exists := current[key]
		if !exists {
			return value
		}

		if i == len(keys)-1 {
			if v, ok := val.(T); ok {
				return v
			}
			return value
		}

		next, ok := val.(map[string]any)
		if !ok {
			return value
		}
		current = next
	}

	return value
}
