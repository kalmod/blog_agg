-- name: CreateFeedFollow :one
insert into feed_follow (id, feed_id, user_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follow WHERE id = $1;

-- name: SelectFeedFollowUser :many
select * from feed_follow where user_id = $1;
