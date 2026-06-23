package auth

import (
	"net/http"
	"testing"
)

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "wrongPassword",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match different hash",
			password:      password1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "invalidhash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	// first we need to create a header
	w := http.Header{}
	w.Set("Authorization", "Bearer MY_TOKEN_STRING")

	tests := []struct {
		name      string
		header    http.Header
		wantErr   bool
		wantToken string
	}{
		{
			name:      "Valid header",
			header:    http.Header{},
			wantErr:   false,
			wantToken: "MY_TOKEN_STRING",
		},
		{
			name:      "Invalid header key",
			header:    http.Header{},
			wantErr:   true,
			wantToken: "",
		},
		{
			name:      "Invalid header value",
			header:    http.Header{},
			wantErr:   true,
			wantToken: "",
		},
		{
			name:      "No authorization header",
			header:    http.Header{},
			wantErr:   true,
			wantToken: "",
		},
	}

	tests[0].header.Set("Authorization", "Bearer MY_TOKEN_STRING")
	tests[1].header.Set("Authorizatio", "Bearer MY_TOKEN_STRING")
	tests[2].header.Set("Authorization", "Bearers MY_TOKEN_STRING")
	// do different scenarios for valid header and non valid,
	// bad written and empty

	// run the test for valid should get a token string and no error
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := GetBearerToken(tt.header)
			if !tt.wantErr && tokenString != "MY_TOKEN_STRING" && err != nil {
				t.Errorf("GetBearerToken() error= %v", err)
			}
		})
	}
}
