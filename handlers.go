package main

import (
	"net/http"
	"strings"
	"time"

	uuid "github.com/google/uuid"
	"github.com/kalmod/blog_agg/internal/database"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	type statusJSON struct {
		Status string `json:"status"`
	}
	respondWithJSON(w, 200, statusJSON{Status: "ok"})
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 500, "Internal Server Error")
}

func (db *apiConfig) postUsersHandler(w http.ResponseWriter, r *http.Request) {
	// unpackage json
	// 1: create decorder, create corresponding struct, decode into struct
	newUser, err := decodeUsers(r)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	createdUser := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      newUser.Name,
	}
	db.DB.CreateUser(r.Context(), createdUser)

	respondWithJSON(w, http.StatusInternalServerError, createdUser)
}

func (db *apiConfig) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	authorizationHeader := r.Header.Get("Authorization")
	api_key := strings.Split(authorizationHeader, " ")[1]
	if api_key == "" {
		respondWithError(w, http.StatusInternalServerError, "No Header")
		return
	}
	userInfo, err := db.DB.GetUserByAPI(r.Context(), api_key)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve from DB")
		return
	}
	respondWithJSON(w, http.StatusOK, userInfo)
	return
}
