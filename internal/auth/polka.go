package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	h := headers.Get("Authorization")

	if h == "" {
		return "", errors.New("No ApiKey provided")
	}

	keyString := strings.TrimPrefix(h, "ApiKey ")

	return keyString, nil
}
