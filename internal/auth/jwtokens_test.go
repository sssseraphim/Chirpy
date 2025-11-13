package auth

import (
	"github.com/google/uuid"
	"net/http"
	"testing"
	"time"
)

func TestTokens(t *testing.T) {
	userId := uuid.New()
	token, _ := MakeJWT(userId, "correct_secret", time.Hour)
	validatedId, err := ValidateJWT(token, "correct_secret")
	if err != nil {
		t.Errorf("coudnt validate: %v", err)
	}
	if userId != validatedId {
		t.Errorf("wrong valid")
	}
	userId = uuid.New()
	token, err = MakeJWT(userId, "secret", 1000*time.Millisecond)
	if err != nil {
		t.Errorf("jit tripin: %v", err)
	}
	validatedId, err = ValidateJWT(token, "secret")
	if err != nil {
		t.Errorf("token should be valid: %v", err)
	}
	time.Sleep(1500 * time.Millisecond)
	_, err = ValidateJWT(token, "secret")
	if err == nil {
		t.Error("token should be invalid")
	}
	header := make(http.Header)
	header.Add("Authorization", "Bearer bombaclat")
	bearerToken, err := GetBearerToken(header)
	if err != nil {
		t.Errorf("failed to get bearer token: %v", err)
	}
	if bearerToken != "bombaclat" {
		t.Error("wrong token_string")
	}
}
