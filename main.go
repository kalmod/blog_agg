package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorJSON struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorJSON{Error: msg})
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	PORT := os.Getenv("PORT")
	mux := http.NewServeMux()

	// Init Handlers
	mux.HandleFunc("GET /v1/healthz", healthHandler)
	mux.HandleFunc("GET /v1/err", errorHandler)

	server := http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	// https://dev.to/mokiat/proper-http-shutdown-in-go-3fji
	go func() {
		log.Println("Starting Server....")
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			// normally returns ErrServerClosed
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped Server.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTPP close error: %v", err)
	}

}
