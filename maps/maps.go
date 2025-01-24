package maps

import (
	"encoding/json"
	"errors"
	"strings"
)

// ToStruct is a helper function to convert map[string]any to struct
func ToStruct(data map[string]any, v any) error {
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

// MapToStruct is a helper function to convert map[string]any to struct
//
// Deprecated: Use [ToStruct] instead.
func MapToStruct(data any, v any) error {
	dd, ok := data.(map[string]any)
	if !ok {
		return errors.New("can not convert data to struct")
	}
	return ToStruct(dd, v)
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

// FromStruct is a helper function to make map[string]any from struct
func FromStruct(v any) (map[string]any, error) {
	jsoned, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	err = json.Unmarshal(jsoned, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
