package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/alborithm/boot-chirpy/internal/auth"
)

func (cfg *apiConfig) HandlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type RequestCred struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	requestCred := RequestCred{}
	err := decoder.Decode(&requestCred)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
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

	userView := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
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
