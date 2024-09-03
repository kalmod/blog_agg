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

	// users
	mux.HandleFunc("POST /v1/users", dbConfig.postUsersHandler)
	mux.HandleFunc("GET /v1/users", dbConfig.middlewareAuth(dbConfig.getUsersHandler))
	// feeds
	mux.HandleFunc("POST /v1/feeds", dbConfig.middlewareAuth(dbConfig.postFeedsHandler))
	mux.HandleFunc("GET /v1/feeds", dbConfig.getAllFeeds)
	// feed_follows
	mux.HandleFunc("POST /v1/feed_follows", dbConfig.middlewareAuth(dbConfig.postFeedFollows))
	mux.HandleFunc("DELETE /v1/feed_follows/{feedFollowID}", dbConfig.deleteFeedFollows)
	mux.HandleFunc("GET /v1/feed_follows", dbConfig.middlewareAuth(dbConfig.getFeedFollowsForUser))

	server := http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	ticker := time.NewTicker(60 * time.Second)
	done := make(chan bool)

	// https://dev.to/mokiat/proper-http-shutdown-in-go-3fji
	go func() {
		log.Println("Starting Server....")
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			// normally returns ErrServerClosed
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped Server.")
	}()

	go func() {
		for {
			select {
			case <-ticker.C:
				dbConfig.FeedWorker(2)
			case <-done:
				return
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()
	ticker.Stop()
	done <- true

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTPP close error: %v", err)
	}

}
