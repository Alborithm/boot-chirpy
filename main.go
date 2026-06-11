package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/alborithm/boot-chirpy/internal/database"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getFileServerHits() int32 {
	return cfg.fileserverHits.Load()
}

func (cfg *apiConfig) resetFileServerHits() {
	cfg.fileserverHits.Swap(0)
}

func main() {
	const port = "8080"
	const filepathRoot = "."
	apiCfg := apiConfig{}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbURL := os.Getenv("DB_URL")

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error opening postgresql connection")
	}
	apiCfg.db = database.New(dbConn)

	mux := http.NewServeMux()

	mux.Handle(
		"/app/",
		apiCfg.middlewareMetricsInc(
			http.StripPrefix(
				"/app",
				http.FileServer(http.Dir(filepathRoot)))))

	mux.HandleFunc("GET /api/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8") // normal header
		w.WriteHeader(http.StatusOK)

		response := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", apiCfg.getFileServerHits())
		w.Write([]byte(response))
	})

	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusOK)

		// response := fmt.Sprintf("Hits: %d", apiCfg.getFileServerHits())
		apiCfg.resetFileServerHits()
		w.Write([]byte("Hits reset"))
	})

	// POST /api/validate_chirp
	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		type chirpPost struct {
			Body string `json:"body"`
		}

		type errorBody struct {
			Error string `json:"error"`
		}

		type responseBody struct {
			CleanedBody string `json:"cleaned_body"`
		}
		w.Header().Set("Content-Type", "application/json")

		decoder := json.NewDecoder(r.Body)
		chirp := chirpPost{}
		err := decoder.Decode(&chirp)
		if err != nil || chirp.Body == "" {
			w.WriteHeader(500)

			errorResponse := errorBody{
				Error: "Something went wrong",
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

		validResponse := responseBody{
			CleanedBody: strings.Join(words, " "),
		}

		dat, err := json.Marshal(validResponse)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
		w.Write(dat)
	})

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		type userRequest struct {
			Email string `json:"email"`
		}

		w.Header().Set("Content-Type", "application/json")

		decoder := json.NewDecoder(r.Body)
		userReq := userRequest{}
		err = decoder.Decode(&userReq)
		if err != nil || userReq.Email == "" {
			w.WriteHeader(500)
			return
		}

		usr, err := apiCfg.db.CreateUser(r.Context(), userReq.Email)
		if err != nil {
			w.WriteHeader(500)
			log.Fatal(err)
			return
		}

		jsonData, err := json.Marshal(usr)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(201)
		w.Write(jsonData)
	})

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
