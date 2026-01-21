package gowok

import (
	"testing"

	"github.com/golang-must/must"
)

func TestHash_Password(t *testing.T) {
	t.Run("positive/generate with salt", func(t *testing.T) {
		raw := "password123"
		salt := "somesalt"
		p := Hash.Password(raw, salt)
		must.Equal(t, salt, p.Salt)
		must.NotEqual(t, "", p.Hashed)
	})

	t.Run("positive/generate without salt", func(t *testing.T) {
		raw := "password123"
		p := Hash.Password(raw)
		must.NotEqual(t, "", p.Salt)
		must.NotEqual(t, "", p.Hashed)
	})
}

func TestHash_PasswordVerify(t *testing.T) {
	testCases := []struct {
		name     string
		raw      string
		password Password
		expected bool
	}{
		{
			name: "positive/correct password",
			raw:  "password123",
			password: func() Password {
				return Hash.Password("password123")
			}(),
			expected: true,
		},
		{
			name: "negative/wrong password",
			raw:  "wrongpassword",
			password: func() Password {
				return Hash.Password("password123")
			}(),
			expected: false,
		},
		{
			name: "negative/wrong salt",
			raw:  "password123",
			password: func() Password {
				p := Hash.Password("password123")
				p.Salt = "wrongsalt"
				return p
			}(),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := Hash.PasswordVerify(tc.raw, tc.password.Hashed, tc.password.Salt)
			must.Equal(t, tc.expected, isValid)
		})
	}
}
