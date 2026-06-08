package main

import (
	"log"
	"net/http"
)

func main() {
	port := "8080"
	filepathRoot := "."

	mux := http.NewServeMux()

	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))

	mux.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("OK"))
	})

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
