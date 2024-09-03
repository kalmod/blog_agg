package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/kalmod/blog_agg/internal/database"
)

// This function should be run every x seconds
func (db *apiConfig) FeedWorker(n int) {
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
			_ = ProcessFeed(feed)
			err := db.DB.MarkFeedFetched(ctx, feed.ID)
			if err != nil {
				fmt.Printf("ERROR: %v - %v", feed.Name, err.Error())
			}
		}()
	}
	wg.Wait()
}

func ProcessFeed(feed database.Feed) error {
	rss, err := GetFeedDataFromUrl(feed.Url)
	if err != nil {
		return err
	}
	fmt.Printf("~~%v~~\n", rss.Channel.Title)
	for _, item := range rss.Channel.Item {
		fmt.Printf("%+v\n", item.Title)
	}

	return nil
}
