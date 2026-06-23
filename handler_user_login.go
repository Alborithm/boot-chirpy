package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alborithm/boot-chirpy/internal/auth"
	"github.com/alborithm/boot-chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type RequestCred struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	requestCred := RequestCred{}
	err := decoder.Decode(&requestCred)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if requestCred.ExpiresInSeconds >= 0 || requestCred.ExpiresInSeconds > (1*60*60) {
		requestCred.ExpiresInSeconds = (1 * 60 * 60)
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), requestCred.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Incorrect email or password"))
		return
	}

	passwordMatch, err := auth.CheckPasswordHash(requestCred.Password, user.HashedPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	if !passwordMatch {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Incorrect email or password"))
		return
	}

	// Create token
	duration, err := time.ParseDuration(fmt.Sprintf("%ds", requestCred.ExpiresInSeconds))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.jwtSecret, duration)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refreshToken, err := CreateRefreshToken(cfg, r, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error creating a refresh token: %v", err)))
	}

	userView := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        jwt,
		RefreshToken: refreshToken,
	}

	dat, err := json.Marshal(userView)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)
}

func CreateRefreshToken(cfg *apiConfig, r *http.Request, userID uuid.UUID) (string, error) {
	// Insert refresh token in database
	resultToken, err := cfg.db.CreateRefreshToken(
		r.Context(),
		database.CreateRefreshTokenParams{
			Token:  auth.MakeRefreshToken(),
			UserID: userID,
		})

	if err != nil {
		return "", err
	}

	// Return refresh token
	return resultToken.Token, nil
}
