-- name: GetSessionOwnerByID :one
SELECT
    id,
    photographer_id
FROM sessions
WHERE id = $1;

-- name: InsertSession :one
INSERT INTO sessions (
  id,
  photographer_id,
  title,
  client_email,
  base_price_cents,
  included_count,
  extra_price_cents,
  min_select_count,
  currency,
  payment_mode

) VALUES(
sqlc.arg(id),
sqlc.arg(photographer_id),
sqlc.arg(title),
sqlc.narg(client_email),
sqlc.arg(base_price_cents),
sqlc.arg(included_count),
sqlc.arg(extra_price_cents),
sqlc.arg(min_select_count),
sqlc.arg(currency),
sqlc.arg(payment_mode)
)
RETURNING 
  id,
  status;

-- name: GetSessions :many
SELECT 
  id,
  photographer_id,
  title,
  client_email,
  status,
  base_price_cents,
  included_count,
  extra_price_cents,
  min_select_count,
  currency,
  payment_mode,
  created_at,
  updated_at,
  closed_at,
  delete_after
FROM sessions
WHERE photographer_id = $1
ORDER BY created_at DESC
LIMIT 200
OFFSET $2;

-- name: GetSessionByID :one
SELECT 
  id,
  photographer_id,
  title,
  client_email,
  status,
  base_price_cents,
  included_count,
  extra_price_cents,
  min_select_count,
  currency,
  payment_mode,
  created_at,
  updated_at,
  closed_at,
  delete_after
FROM sessions
WHERE id = $1;

-- name: CloseSession :one
UPDATE sessions
SET
  status = 'closed',
  closed_at = CURRENT_TIMESTAMP,
  delete_after = CURRENT_TIMESTAMP + INTERVAL '90 days'
WHERE id = $1
RETURNING
  id,
  title,
  status,
  closed_at,
  delete_after;

-- name: MarkSessionProcessing :execrows
UPDATE sessions
SET
    status = 'processing',
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND status = 'draft';

-- name: MarkSessionSelecting :execrows
UPDATE sessions
SET
    status = 'selecting',
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND status = 'processing';

-- name: MarkSessionFailed :execrows
UPDATE sessions
SET
    status = 'failed',
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND status IN ('draft', 'processing');

-- name: GetSessionStatusForUpdate :one
SELECT
  id,
  status
FROM sessions
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: MarkSessionEditing :execrows
UPDATE sessions
SET
  status = 'editing',
  updated_at = now()
WHERE id = sqlc.arg(id)
  AND status = 'waiting_for_payment';
