-- name: CreateFeed :one
INSERT INTO feeds (id,created_at,updated_at, name, url, user_id, last_fetched_at)
VALUES ($1, $2, $3, $4, $5, $6, NULL)
RETURNING *;

-- name: SelectAllFeeds :many
SELECT * FROM feeds
ORDER BY created_at;

-- name: GetNextFeedsToFetch :many
SELECT * FROM feeds 
ORDER BY last_fetched_at NULLS FIRST LIMIT $1;

-- name: MarkFeedFetched :exec
UPDATE feeds SET last_fetched_at = LOCALTIMESTAMP WHERE id = $1;
