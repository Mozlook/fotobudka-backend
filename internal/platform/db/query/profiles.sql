-- name: GetPhotographerProfileByUserID :one
SELECT
    user_id,
    username,
    display_name,
    bio,
    social_links,
    created_at,
    updated_at
FROM photographer_profiles
WHERE user_id = $1;

-- name: UpsertPhotographerProfile :one
INSERT INTO photographer_profiles (
    user_id,
    username,
    display_name,
    bio,
    social_links
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(username),
    sqlc.arg(display_name),
    sqlc.arg(bio),
    sqlc.arg(social_links)
)
ON CONFLICT (user_id) DO UPDATE
SET
    username = EXCLUDED.username,
    display_name = EXCLUDED.display_name,
    bio = EXCLUDED.bio,
    social_links = EXCLUDED.social_links,
    updated_at = now()
RETURNING
    user_id,
    username,
    display_name,
    bio,
    social_links,
    created_at,
    updated_at;
