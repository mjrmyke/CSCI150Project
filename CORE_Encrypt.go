package main

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
)

func Encrypt(data []byte, key []byte) (string, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	for len(data) < b.BlockSize() {
		data = append(data, '=')
	}
	res := make([]byte, len(data))
	b.Encrypt(res, data)
	finalValue := base64.StdEncoding.EncodeToString(res)
	return finalValue, nil
}

func Decrypt(data string, key []byte) ([]byte, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	strData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	res := make([]byte, len(strData))
	b.Decrypt(res, strData)
	return bytes.TrimRight(res, "="), nil
}
