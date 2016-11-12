// A set of functions for making and maintaining cookies.
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
)

func createHmac(value string) ([]byte, error) {
	mac := hmac.New(sha256.New, []byte(HMAC_Key))
	_, err := io.WriteString(mac, value)
	if err != nil {
		return []byte{}, err
	}
	return mac.Sum(nil), nil
}

func splitMac(value string) (string, string) {
	i := strings.LastIndex(value, ".")
	if i == -1 {
		return value, ""
	}
	return value[:i], value[i+1:]
}

func checkMac(value, mac string) bool {
	derivedMac, err := createHmac(value)
	if err != nil {
		return false
	}
	macData, err := base64.RawURLEncoding.DecodeString(mac)
	if err != nil {
		return false
	}
	return hmac.Equal(derivedMac, macData)
}

func DeleteCookie(res http.ResponseWriter, name string) {
	http.SetCookie(res, &http.Cookie{
		Name:   name,
		MaxAge: -1,
		Path:   "/",
	})
}

func MakeCookie(res http.ResponseWriter, name, value string) error {
	mac, err := createHmac(value)
	if err != nil {
		return err
	}
	c := &http.Cookie{
		Name:     name,
		Value:    value + "." + base64.RawURLEncoding.EncodeToString(mac),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   sessionTime,
	}
	http.SetCookie(res, c)
	return nil
}

func GetCookieValue(req *http.Request, name string) (string, error) {
	cookie, err := req.Cookie(name)
	if err != nil {
		return "", err
	}
	val, mac := splitMac(cookie.Value)
	if good := checkMac(val, mac); !good {
		return "", ErrNotMatchingHMac
	}
	return val, nil
}
