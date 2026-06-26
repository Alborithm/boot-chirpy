package main

import (
	"encoding/json"
	"net/http"
)

func respondWithError(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	type errorBody struct {
		Error string `json:"error"`
	}

	w.WriteHeader(statusCode)

	errorResponse := errorBody{
		Error: message,
	}

	dat, err := json.Marshal(errorResponse)
	if err != nil {
		return
	}
	w.Write(dat)
}
