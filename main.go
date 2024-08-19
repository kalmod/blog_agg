package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kalmod/blog_agg/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("ENV Error: %v\n", err)
	}
	PORT := os.Getenv("PORT")
	DBURL := os.Getenv("DB_CONNECTION")

	db, err := sql.Open("postgres", DBURL)
	if err != nil {
		log.Fatalf("DB Error: %v\n", err)
	}

	dbQueries := database.New(db)
	dbConfig := apiConfig{DB: dbQueries}

	mux := http.NewServeMux()

	// Init Handlers
	mux.HandleFunc("GET /v1/healthz", healthHandler)
	mux.HandleFunc("GET /v1/err", errorHandler)

	mux.HandleFunc("POST /v1/users", dbConfig.postUsersHandler)
	mux.HandleFunc("GET /v1/users", dbConfig.getUsersHandler)

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
