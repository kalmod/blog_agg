-- +goose Up
CREATE TABLE feed_follow (
	id UUID PRIMARY KEY,
	feed_id UUID NOT NULL REFERENCES feeds ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
  FOREIGN KEY (feed_id) REFERENCES feeds (id),
  FOREIGN KEY (user_id) REFERENCES users (id)
);

-- +goose Down
DROP TABLE feed_follow;
