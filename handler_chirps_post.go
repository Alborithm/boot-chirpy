package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/alborithm/boot-chirpy/internal/auth"
	"github.com/alborithm/boot-chirpy/internal/database"
)

func (cfg *apiConfig) handlerChirpsPost(w http.ResponseWriter, r *http.Request) {
	type chirpPost struct {
		Body string `json:"body"`
	}

	type errorBody struct {
		Error string `json:"error"`
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	chirp := chirpPost{}
	err := decoder.Decode(&chirp)
	if err != nil || chirp.Body == "" {
		w.WriteHeader(500)

		errorResponse := errorBody{
			Error: fmt.Sprintf("Chirp could not be decoded to JSON object: %v", err),
		}

		dat, err := json.Marshal(errorResponse)
		if err != nil {
			return
		}
		w.Write(dat)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		errorResponse := errorBody{
			Error: fmt.Sprintf("Could not get token: %v", err),
		}

		dat, err := json.Marshal(errorResponse)
		if err != nil {
			return
		}
		w.Write(dat)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)

		errorResponse := errorBody{
			Error: fmt.Sprintf("Auth Error: %v", err),
		}

		dat, err := json.Marshal(errorResponse)
		if err != nil {
			return
		}
		w.Write(dat)
		return
	}

	// Long chirp check
	if len(chirp.Body) > 140 {
		w.WriteHeader(400)

		errorResponse := errorBody{
			Error: "Chirp is too long",
		}

		dat, err := json.Marshal(errorResponse)
		if err != nil {
			return
		}

		w.Write(dat)
		return
	}

	// Cleane the input

	words := strings.Split(chirp.Body, " ")

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	for i, word := range words {
		if _, ok := badWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}

	chirpRequest := database.CreateChirpParams{
		Body:   strings.Join(words, " "),
		UserID: userID,
	}

	response, err := cfg.db.CreateChirp(r.Context(), chirpRequest)
	if err != nil {
		w.WriteHeader(500)
		errorResponse := errorBody{
			Error: fmt.Sprintf("Error inserting chirp in database: %v", err),
		}

		dat, err := json.Marshal(errorResponse)
		if err != nil {
			return
		}

		w.Write(dat)
		return
	}

	newChirp := Chirp{
		ID:        response.ID,
		CreatedAt: response.CreatedAt,
		UpdatedAt: response.UpdatedAt,
		Body:      response.Body,
		UserID:    response.UserID,
	}

	dat, err := json.Marshal(newChirp)
	if err != nil {
		w.WriteHeader(500)

		errorResponse := errorBody{
			Error: fmt.Sprintf("Error marshalling chirp: %v", err),
		}

		dat, err := json.Marshal(errorResponse)
		if err != nil {
			return
		}
		w.Write(dat)

		return
	}
	w.WriteHeader(201)
	w.Write(dat)
}
