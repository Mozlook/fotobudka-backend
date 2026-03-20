-- name: GetSessionOwnerByID :one
SELECT
    id,
    photographer_id
FROM sessions
WHERE id = $1;


