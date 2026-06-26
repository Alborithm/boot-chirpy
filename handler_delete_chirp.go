package main

import (
	"fmt"
	"net/http"

	"github.com/alborithm/boot-chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, r, http.StatusUnauthorized, fmt.Sprintf("Authirization error : %v", err))
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, r, http.StatusUnauthorized, fmt.Sprintf("Authirization error : %v", err))
		return
	}

	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, fmt.Sprintf("Error parsing the chirpID: %v", err))
		return
	}
	chirp, err := cfg.db.GetChirpByID(r.Context(), id)
	if err != nil {
		respondWithError(w, r, http.StatusNotFound, fmt.Sprintf("Chirp not found: %v", err))
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, r, http.StatusForbidden, fmt.Sprintf("Unauthorized, this is not your chirp : %v", err))
		return
	}

	err = cfg.db.DeleteChirpByID(r.Context(), id)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, fmt.Sprintf("Error deleting the chirp: %v", err))
	}

	w.WriteHeader(http.StatusNoContent)
}
