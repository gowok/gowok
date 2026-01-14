package file

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

func ExtentionToMime(extension string) string {
	return MimeTypes[extension]
}

func MimeToExtension(mime string) string {
	var foundMime string
	for k, v := range MimeTypes {
		if v == mime {
			foundMime = k
		}
	}

	return foundMime
}

func GetTypeFromBase64(base64encoded string) string {
	_fileType := base64encoded[0:strings.Index(base64encoded, ";")]
	fileType := _fileType[strings.Index(_fileType, "/")+1:]
	return fileType
}

func GetMimeFromBase64(base64encoded string) string {
	return MimeTypes[fmt.Sprintf(".%s", GetTypeFromBase64(base64encoded))]
}

func SaveBase64StringToFile(path string, fileNameWithoutType string, encodedBase64 string) (string, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return "", err
		}
	}

	fileType := GetTypeFromBase64(encodedBase64)
	filePath := path + "/" + fileNameWithoutType + "." + fileType
	encodedFileData := encodedBase64[strings.Index(encodedBase64, ",")+1:]

	fileBase64Decoded := base64.NewDecoder(base64.StdEncoding, strings.NewReader(encodedFileData))

	fileCreated, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = fileCreated.Close()
	}()

	_, err = io.Copy(fileCreated, fileBase64Decoded)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
