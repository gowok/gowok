package gowok

import (
	"crypto/rand"
	"crypto/sha512"
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

func equal(cipherText, newCipherText string) bool {
	x, _ := base64.StdEncoding.DecodeString(cipherText)
	diff := uint64(len(x)) ^ uint64(len(newCipherText))

	for i := 0; i < len(x) && i < len(newCipherText); i++ {
		diff |= uint64(x[i]) ^ uint64(newCipherText[i])
	}

	return diff == 0
}

func (p *_hash) PasswordVerify(raw, hashed, salt string) bool {
	dk := pbkdf2.Key([]byte(raw), []byte(salt), iter, keyLen, sha512.New)

	return equal(hashed, string(dk))
}
