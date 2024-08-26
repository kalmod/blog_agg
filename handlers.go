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
	db.DB.CreateUser(r.Context(), createdUser)

	respondWithJSON(w, http.StatusInternalServerError, createdUser)
}

func (db *apiConfig) getUsersHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, http.StatusOK, user)
}

func (db *apiConfig) postFeedsHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	type feedsPost struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	feedData, err := myDecoder[feedsPost](r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Error")
	}
	createdFeed := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      feedData.Name,
		Url:       feedData.URL,
		UserID:    user.ID,
	}

	result, err := db.DB.CreateFeed(r.Context(), createdFeed)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid Request")
	}
	respondWithJSON(w, http.StatusOK, result)
}

func (db *apiConfig) getAllFeeds(w http.ResponseWriter, r *http.Request) {

	allFeeds, err := db.DB.SelectAllFeeds(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "No response from Server")
		return
	}
	respondWithJSON(w, http.StatusOK, allFeeds)
}

func (db *apiConfig) postFeedFollows(w http.ResponseWriter, r *http.Request, user database.User) {
	type feed_follow_post struct {
		Feed_Id uuid.UUID `json:"feed_id"`
	}
	feedID, err := myDecoder[feed_follow_post](r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	newFeedFollow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		FeedID:    feedID.Feed_Id,
		UserID:    user.ID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	result, err := db.DB.CreateFeedFollow(r.Context(), newFeedFollow)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Error")
		return
	}
	respondWithJSON(w, http.StatusOK, result)
}

func (db *apiConfig) deleteFeedFollows(w http.ResponseWriter, r *http.Request) {
	pathSlice := strings.Split(r.URL.Path, "/")
	feedId := pathSlice[len(pathSlice)-1]
	feedUUID, err := uuid.FromBytes([]byte(feedId))
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, "Internal Error")
	}
	db.DB.DeleteFeedFollow(r.Context(), feedUUID)
	respondWithJSON(w, http.StatusOK, struct {
		Body string `json:"body"`
	}{fmt.Sprintf("%v: Deleted Feed Follow entries", feedId)})
}

func (db *apiConfig) getFeedFollowsForUser(w http.ResponseWriter, r *http.Request, user database.User) {
	result, err := db.DB.SelectFeedFollowUser(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Error")
		return
	}
	respondWithJSON(w, http.StatusOK, result)
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
