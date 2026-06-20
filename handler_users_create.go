package main

import (
	"encoding/json"
	"net/http"

	"github.com/alborithm/boot-chirpy/internal/auth"
	"github.com/alborithm/boot-chirpy/internal/database"
)

func (cfg *apiConfig) HandlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type userRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	userReq := userRequest{}
	err := decoder.Decode(&userReq)
	if err != nil || userReq.Email == "" || userReq.Password == "" {
		w.WriteHeader(500)
		return
	}

	hashedPassword, err := auth.HashPassword(userReq.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error hashing the password: " + err.Error()))
		return
	}

	response, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          userReq.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {
		w.WriteHeader(500)
		return
	}

	usr := User{
		ID:        response.ID,
		CreatedAt: response.CreatedAt,
		UpdatedAt: response.UpdatedAt,
		Email:     response.Email,
	}

	jsonData, err := json.Marshal(usr)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(201)
	w.Write(jsonData)
}
