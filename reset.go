package main

import "net/http"

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		w.Write([]byte("Reset is only allowed in dev environment."))
		return
	}

	cfg.resetFileServerHits()

	if err := cfg.db.DeleteUsers(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to reset the database: " + err.Error()))
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0 and database reset to inital state."))
}

func (cfg *apiConfig) resetFileServerHits() {
	cfg.fileserverHits.Swap(0)
}
