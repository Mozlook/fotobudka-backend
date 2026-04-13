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

-- name: GetClientSessionByTokenHMAC :one
SELECT 
  session_access.id AS session_access_id,
  sessions.id AS session_id,
  sessions.status,
  sessions.base_price_cents,
  sessions.included_count,
  sessions.extra_price_cents,
  sessions.min_select_count,
  sessions.currency,
  sessions.payment_mode,
  sessions.title
FROM session_access
JOIN sessions ON session_access.session_id = sessions.id
WHERE session_access.token_hmac = $1
AND session_access.revoked_at IS NULL;


-- name: GetClientSessionByCodeHMAC :one
SELECT 
  session_access.id AS session_access_id,
  sessions.id AS session_id,
  sessions.status,
  sessions.base_price_cents,
  sessions.included_count,
  sessions.extra_price_cents,
  sessions.min_select_count,
  sessions.currency,
  sessions.payment_mode,
  sessions.title
FROM session_access
JOIN sessions ON session_access.session_id = sessions.id
WHERE session_access.code_hmac = $1
AND session_access.revoked_at IS NULL;

-- name: GetActiveClientSessionAccessByID :one
SELECT
  id,
  session_id
FROM session_access
WHERE id = sqlc.arg(id)
AND revoked_at IS NULL;
