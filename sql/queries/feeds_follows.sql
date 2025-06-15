-- name: CreateFeedFollow :exec
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
        VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
select
    inserted_feed_follow.*,
    u.name AS feed_name,
    f.name AS user_name
from inserted_feed_follow
inner join
    users u on inserted_feed_follow.user_id = u.id
inner join
    feeds f on inserted_feed_follow.feed_id = f.id;

-- name: GetFeedFollowsForUser :many
select f.name as feed_name, u.name as user_name
from feed_follows
inner join
    users u on feed_follows.user_id = u.id
inner join
    feeds f on feed_follows.feed_id = f.id
where feed_follows.user_id = $1;

-- name: ResetFeedFollow :exec
DELETE FROM feed_follows;
