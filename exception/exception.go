package exception

import (
	"errors"
	"fmt"
)

var (
	ConfigNotFound = errors.New("config file not found")
	ConfigDecoding = func(err error) error { return fmt.Errorf("config decoding failed: %s", err.Error()) }
)
