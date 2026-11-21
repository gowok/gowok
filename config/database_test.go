package config

import (
	"testing"
	"time"

	"github.com/golang-must/must"
)

func TestSQL_GetConnMaxLifetime(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		expected time.Duration
	}{
		{
			name:     "positive/custom duration",
			value:    "5m",
			expected: 5 * time.Minute,
		},
		{
			name:     "positive/empty uses default",
			value:    "",
			expected: 0,
		},
		{
			name:     "negative/invalid uses default",
			value:    "invalid",
			expected: 0,
		},
		{
			name:     "negative/non-positive uses default",
			value:    "-5m",
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := SQL{ConnMaxLifetime: tc.value}
			must.Equal(t, tc.expected, cfg.GetConnMaxLifetime())
		})
	}
}

func TestSQL_GetConnMaxIdleTime(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		expected time.Duration
	}{
		{
			name:     "positive/custom duration",
			value:    "30s",
			expected: 30 * time.Second,
		},
		{
			name:     "positive/empty uses default",
			value:    "",
			expected: 0,
		},
		{
			name:     "negative/invalid uses default",
			value:    "invalid",
			expected: 0,
		},
		{
			name:     "negative/non-positive uses default",
			value:    "-30s",
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := SQL{ConnMaxIdleTime: tc.value}
			must.Equal(t, tc.expected, cfg.GetConnMaxIdleTime())
		})
	}
}
