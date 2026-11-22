package gowok

import (
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"

	"golang.org/x/crypto/pbkdf2"
)

type _hash struct{}

var Hash = _hash{}

type Password struct {
	Hashed string
	Salt   string
}

var iter = 10000
var keyLen = 32

func genSalt(size int) string {
	saltBytes := make([]byte, size)
	_, err := rand.Read(saltBytes)
	if err != nil {
		panic(err)
	}

	result := base64.StdEncoding.EncodeToString(saltBytes)
	return result
}

func (p *_hash) Password(raw string, salt ...string) Password {
	if len(salt) <= 0 {
		salt = []string{genSalt(16)}
	}

	dk := pbkdf2.Key([]byte(raw), []byte(salt[0]), iter, keyLen, sha512.New)
	dkString := base64.StdEncoding.EncodeToString(dk)
	return Password{
		Hashed: dkString,
		Salt:   salt[0],
	}
}

func (p *_hash) PasswordVerify(raw, hashed, salt string) bool {
	stored, err := base64.StdEncoding.DecodeString(hashed)
	if err != nil {
		return false
	}

	dk := pbkdf2.Key([]byte(raw), []byte(salt), iter, keyLen, sha512.New)

	return subtle.ConstantTimeCompare(stored, dk) == 1
}
