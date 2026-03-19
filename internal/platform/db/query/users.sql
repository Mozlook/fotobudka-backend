-- name: UpsertUserFromGoogle :one
INSERT INTO users (
    id,
    google_sub,
    email,
    name,
    avatar_url
) VALUES (
    sqlc.arg(id),
    sqlc.arg(google_sub),
    sqlc.arg(email),
    sqlc.arg(name),
    NULLIF(sqlc.arg(avatar_url), '')
)
ON CONFLICT (google_sub) DO UPDATE
SET
    email = EXCLUDED.email,
    name = EXCLUDED.name,
    avatar_url = EXCLUDED.avatar_url
RETURNING
    id,
    google_sub,
    email,
    name,
    avatar_url,
    created_at;
