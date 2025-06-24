-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetAllFeeds :many
SELECT f.name, f.url, u.name
FROM feeds f
JOIN users u ON f.user_id = u.id;

-- name: GetFeed :one
SELECT id
FROM feeds
WHERE url = $1;

-- name: ResetFeeds :exec
DELETE FROM feeds;

-- name: MarkFeedFetched :one
UPDATE feeds
SET last_fetched_at = NOW(),
updated_at = NOW()
WHERE id = $1
returning *;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;
