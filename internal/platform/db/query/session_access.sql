-- name: InsertSessionAccess :one
INSERT INTO session_access (
  session_id,
  code_hmac,
  token_hmac
) VALUES (
sqlc.arg(session_id),
sqlc.arg(code_hmac),
sqlc.arg(token_hmac)
)
RETURNING
  id,
  created_at;
