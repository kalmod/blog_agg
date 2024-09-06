package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kalmod/blog_agg/internal/database"
)

type Posts struct {
	ID           uuid.UUID
	created_at   time.Time
	updated_at   time.Time
	title        string
	url          string
	description  string
	published_at time.Time
	feed_id      uuid.UUID
}

// This function should be run every x seconds
func (db *apiConfig) FeedWorker(n int) {
	fmt.Println("Pulling posts from listed feeds....")
	ctx := context.Background()
	feeds, err := db.DB.GetNextFeedsToFetch(ctx, int32(n))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var wg sync.WaitGroup

	for _, feed := range feeds {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = db.ProcessFeed(ctx, feed)
			err := db.DB.MarkFeedFetched(ctx, feed.ID)
			if err != nil {
				fmt.Printf("ERROR: %v - %v", feed.Name, err.Error())
			}
		}()
	}
	wg.Wait()
}

func toNullString(text string) sql.NullString {
	if text == "" {
		return sql.NullString{String: "", Valid: false}
	}
	return sql.NullString{String: text, Valid: true}
}

func toNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Time: t, Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func (db *apiConfig) ProcessFeed(ctx context.Context, feed database.Feed) error {
	rss, err := GetFeedDataFromUrl(feed.Url)
	if err != nil {
		return err
	}

	for _, item := range rss.Channel.Item {
		pubTime, err := time.Parse(time.RFC1123, item.PubDate)
		if err != nil {
			fmt.Printf("Couldn't parse %v", item.PubDate)
			return err
		}
		newPost := database.CreatePostsParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Url:         item.Link,
			Description: toNullString(item.Description),
			PublishedAt: toNullTime(pubTime),
			FeedID:      feed.ID,
		}
		_, err = db.DB.CreatePosts(ctx, newPost)
		if err != nil {
			fmt.Printf("CreatePosts Error: %v\n", err.Error())
			return err
		}
	}

	return nil
}
