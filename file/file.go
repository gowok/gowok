package file

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
