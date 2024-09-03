package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	uuid "github.com/google/uuid"
	"github.com/kalmod/blog_agg/internal/database"
)

type Feed struct {
	ID            uuid.UUID  `json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Name          string     `json:"name"`
	Url           string     `json:"url"`
	UserID        uuid.UUID  `json:"user_id"`
	LastFetchedAt *time.Time `json:"last_fetched_at"`
}

func databaseFeedToFeed(feed database.Feed) Feed {
	return Feed{
		ID:            feed.ID,
		CreatedAt:     feed.CreatedAt,
		UpdatedAt:     feed.UpdatedAt,
		Name:          feed.Name,
		Url:           feed.Url,
		UserID:        feed.UserID,
		LastFetchedAt: &feed.LastFetchedAt.Time,
	}
}

func (db *apiConfig) getAllFeeds(w http.ResponseWriter, r *http.Request) {
	allFeeds, err := db.DB.SelectAllFeeds(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "No response from Server")
		return
	}

	convertedAllFeeds := []Feed{}
	for _, v := range allFeeds {
		convertedAllFeeds = append(convertedAllFeeds, databaseFeedToFeed(v))
	}

	respondWithJSON(w, http.StatusOK, convertedAllFeeds)
}

func (db *apiConfig) postFeedsHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	type feedsPost struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	feedData, err := myDecoder[feedsPost](r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error Feed Decode")
		return
	}
	createdFeed := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      feedData.Name,
		Url:       feedData.URL,
		UserID:    user.ID,
	}

	newFeedFollow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		FeedID:    createdFeed.ID,
		UserID:    user.ID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	feedResult, err := db.DB.CreateFeed(r.Context(), createdFeed)
	if err != nil {
		fmt.Println(err.Error())
		respondWithError(w, http.StatusUnauthorized, "Invalid Request - Create Feed")
		return
	}

	feedFollowResult, feedErr := db.DB.CreateFeedFollow(r.Context(), newFeedFollow)
	if feedErr != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid Request - Create FeedFollow")
		return
	}
	respondWithJSON(w, http.StatusOK,
		struct {
			Feed        Feed                `json:"feed"`
			Feed_Follow database.FeedFollow `json:"feed_follow"`
		}{Feed: databaseFeedToFeed(feedResult), Feed_Follow: feedFollowResult})
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
	feedUUID, err := uuid.Parse(feedId)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	db.DB.DeleteFeedFollow(r.Context(), feedUUID)
	respondWithJSON(w, http.StatusOK, struct {
		Body string `json:"body"`
	}{fmt.Sprintf("%v: Deleted Feed Follow entries", feedId)})
}

func (db *apiConfig) getFeedFollowsForUser(w http.ResponseWriter, r *http.Request, user database.User) {
	result, err := db.DB.SelectFeedFollowUser(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, result)
}
