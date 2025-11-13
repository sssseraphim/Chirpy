package auth

import (
	"fmt"
	"net/http"
)

func GetApiKey(headers http.Header) (string, error) {
	auth_key := headers.Get("Authorization")
	if auth_key == "" {
		return "", fmt.Errorf("no api key")
	}
	return auth_key[7:], nil
}
