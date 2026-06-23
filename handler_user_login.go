package main

import (
	"database/sql"
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

func (cfg *apiConfig) handleUserTokenRefresh(w http.ResponseWriter, r *http.Request) {
	type responseBody struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, fmt.Sprintf("Error on getting bearer token: %v", err))
		return
	}

	tokenRecord, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, r, http.StatusUnauthorized, fmt.Sprintf("Token not found: %v", err))
		return
	}
	if tokenRecord.Token == "" {
		respondWithError(w, r, http.StatusUnauthorized, "Token does not exist")
		return
	}
	nullTime := sql.NullTime{}
	if tokenRecord.RevokedAt != nullTime {
		respondWithError(w, r, http.StatusUnauthorized, "Token revoked")
		return
	}
	if tokenRecord.ExpiresAt.Before(time.Now()) {
		respondWithError(w, r, http.StatusUnauthorized, "Token expired")
		return
	}

	userID, err := cfg.db.GetUserFromRefreshToken(r.Context(), tokenRecord.Token)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, fmt.Sprintf("Error getting the user from token: %v", err))
		return
	}

	jwt, err := auth.MakeJWT(userID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, fmt.Sprintf("Token expired: %v", err))
		return
	}

	response := responseBody{
		Token: jwt,
	}

	dat, err := json.Marshal(response)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, fmt.Sprintf("Error marshalling the token: %v", err))
		return
	}
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) hanldeUserTokenRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, fmt.Sprintf("Error on getting bearer token: %v", err))
		return
	}

	_, err = cfg.db.UpdateRefreshTokenRevoke(r.Context(), token)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, fmt.Sprintf("Error revoking session: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
