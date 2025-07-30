package gowok

import (
	"encoding/json"
	"os"
)

func ToPtr[T any](v T) *T {
	return &v
}

func DD(input any) error {
	return json.NewEncoder(os.Stdout).Encode(input)
}
