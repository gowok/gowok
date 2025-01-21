package gowok

import (
	"errors"
	"os"
	"testing"

	"github.com/gowok/should"
)

func GetTest(config string) (*Project, error) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tempFile, err := os.CreateTemp("", "testIgniteStart*.yaml")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(config)); err != nil {
		return nil, err
	}
	defer tempFile.Close()

	os.Args = []string{"cmd", "--config=" + tempFile.Name(), "--env-file="}
	p := Get()
	if p == nil {
		return nil, errors.New("failed to ignite")
	}
	return p, nil
}

func TestGet(t *testing.T) {
	p, err := GetTest("")
	should.Nil(t, err)
	should.NotNil(t, p)
}
