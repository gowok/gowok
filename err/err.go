package err

import (
	"errors"
	"fmt"
)

var (
	ErrConfigNotFound = errors.New("config file not found")
	ErrConfigDecoding = func(err error) error { return fmt.Errorf("config decoding failed: %s", err.Error()) }
)
