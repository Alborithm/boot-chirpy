package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	// UUID samples
	userID := uuid.New()
	validToken, _ := MakeJWT(userID, "secret", time.Hour)

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantId      uuid.UUID
		wantErr     bool
	}{
		{
			name:        "valid token",
			tokenString: validToken,
			tokenSecret: "secret",
			wantId:      userID,
			wantErr:     false,
		},
		{
			name:        "invalid token",
			tokenString: "invalid.token.string",
			tokenSecret: "secret",
			wantId:      uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong_secret",
			wantId:      uuid.Nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
			}
			if id != tt.wantId {
				t.Errorf("ValidateJWT() expects = %v, got %v", tt.wantId, id)
			}
		})
	}
}
