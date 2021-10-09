package utilities

import (
	"bytes"
	"encoding/base64"
)

func ExportAsBytes(insta *instag.Instagram) ([]byte, error) {
	buffer := &bytes.Buffer{}
	err := insta.Export(insta, buffer)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func ExportAsBase64String(insta *instag.Instagram) (string, error) {
	bytes, err := ExportAsBytes(insta)
	if err != nil {
		return "", err
	}

	sEnc := base64.StdEncoding.EncodeToString(bytes)
	return sEnc, nil
}
