package utilities

import (
	"bytes"
	"encoding/base64"
)


func ImportFromBytes(inputBytes []byte) (*instag.Instagram, error) {
	return instag.ImportReader(bytes.NewReader(inputBytes))
}
func ImportFromBase64String(base64String string) (*instag.Instagram, error) {
	sDec, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return nil, err
	}

	return ImportFromBytes(sDec)
}