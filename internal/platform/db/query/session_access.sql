-- name: InsertSessionAccess :one
INSERT INTO session_access (
  id,
  session_id,
  code_hmac,
  token_hmac
) VALUES (
sqlc.arg(id),
sqlc.arg(session_id),
sqlc.arg(code_hmac),
sqlc.arg(token_hmac)
)
RETURNING
  id,
  created_at;

-- name: RevokeSessionAccess :many
UPDATE session_access
SET revoked_at = now()
WHERE session_id = $1
AND revoked_at IS NULL
RETURNING
  id,
  session_id,
  created_at,
  revoked_at,
  last_used_at;
