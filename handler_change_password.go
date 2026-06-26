package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alborithm/boot-chirpy/internal/auth"
	"github.com/alborithm/boot-chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChangePassword(w http.ResponseWriter, r *http.Request) {
	type RequestParams struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type ResponseBody struct {
		ID        uuid.UUID `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	decoder := json.NewDecoder(r.Body)
	requestParams := RequestParams{}
	err := decoder.Decode(&requestParams)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, fmt.Sprintf("Error decoding json: %v", err))
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, r, http.StatusUnauthorized, fmt.Sprintf("Error getting the token %v", err))
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, r, http.StatusUnauthorized, fmt.Sprintf("Auth error: %v", err))
		return
	}

	hashedPassword, err := auth.HashPassword(requestParams.Password)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, fmt.Sprintf("Error hashing the password %v", err))
		return
	}

	editedUser, err := cfg.db.UpdatePassword(r.Context(), database.UpdatePasswordParams{
		Email:          requestParams.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	})
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, fmt.Sprintf("Error updating password: %v", err))
		return
	}

	response := ResponseBody{
		ID:        editedUser.ID,
		Email:     editedUser.Email,
		CreatedAt: editedUser.CreatedAt,
		UpdatedAt: editedUser.UpdatedAt,
	}
	w.WriteHeader(http.StatusOK)
	dat, err := json.Marshal(response)
	if err != nil {
		return
	}

	w.Write(dat)
}
