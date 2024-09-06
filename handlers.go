package main

import (
	"fmt"
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
	result, err := db.DB.CreateUser(r.Context(), createdUser)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating users")
	}

	respondWithJSON(w, http.StatusOK, result)
}

func (db *apiConfig) getUsersHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, http.StatusOK, user)
}

// middleware custom type
type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (db *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			authorizationHeader := r.Header.Get("Authorization")
			if authorizationHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "Invalid APIKey")
				return
			}
			api_key := strings.Split(authorizationHeader, " ")[1]
			userInfo, err := db.DB.GetUserByAPI(r.Context(), api_key)
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, "Invalid APIKey")
				return
			}
			handler(w, r, userInfo)
		})
}

// Posts
func (db *apiConfig) getPosts(w http.ResponseWriter, r *http.Request, user database.User) {
	type limitJSON struct {
		Limit int `json:"limit"`
	}
	var limitInfo limitJSON
	limitInfo, err := myDecoder[limitJSON](r)
	if err != nil {
		fmt.Println(err.Error())
		limitInfo = limitJSON{Limit: 5}
	}
	posts, err := db.DB.GetPosts(r.Context(), database.GetPostsParams{UserID: user.ID, Limit: int32(limitInfo.Limit)})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving posts")
		return
	}
	respondWithJSON(w, http.StatusOK, posts)
}
