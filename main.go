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
	"time"

	"github.com/alborithm/boot-chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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
	apiCfg.platform = os.Getenv("PLATFORM")

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

	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	// POST /api/chirps
	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		type chirpPost struct {
			Body   string    `json:"body"`
			UserID uuid.UUID `json:"user_id"`
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

		chirpRequest := database.CreateChirpParams{
			Body:   strings.Join(words, " "),
			UserID: chirp.UserID,
		}

		response, err := apiCfg.db.CreateChirp(r.Context(), chirpRequest)
		if err != nil {
			w.WriteHeader(500)
			errorResponse := errorBody{
				Error: "Error inserting chirp",
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
			return
		}
		w.WriteHeader(201)
		w.Write(dat)
	})

	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsGetAll)

	mux.HandleFunc("POST /api/users", apiCfg.HandlerUsersCreate)

	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpsGetByID)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
