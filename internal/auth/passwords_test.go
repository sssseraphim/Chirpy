package auth

import "testing"

func TestHashing(t *testing.T) {
	password := "bob228"
	hash, err := HashPassword(password)
	if err != nil {
		t.Error(err)
	}
	res, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Error(err)
	}
	if !res {
		t.Error("wrong hash")
	}
}
