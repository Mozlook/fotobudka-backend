-- name: GetNextDeliveryVersionForSession :one
SELECT COALESCE(MAX(version), 0)::bigint + 1 AS next_version
FROM deliveries
WHERE session_id = sqlc.arg(session_id);

-- name: InsertDelivery :exec
INSERT INTO deliveries (
  id,
  session_id,
  version,
  status,
  created_at
) VALUES (
  sqlc.arg(id),
  sqlc.arg(session_id),
  sqlc.arg(version),
  sqlc.arg(status),
  now()
);

-- name: GetDeliveryByID :one
SELECT
  id,
  session_id,
  version,
  status,
  zip_key,
  zip_size_bytes,
  created_at,
  generated_at
FROM deliveries
WHERE id = sqlc.arg(id);

-- name: MarkDeliveryReady :execrows
UPDATE deliveries
SET
  status = 'ready',
  zip_key = sqlc.arg(zip_key),
  zip_size_bytes = sqlc.arg(zip_size_bytes),
  generated_at = now()
WHERE id = sqlc.arg(id)
  AND status = 'generating';

-- name: MarkDeliveryFailed :execrows
UPDATE deliveries
SET
  status = 'failed'
WHERE id = sqlc.arg(id)
  AND status = 'generating';

-- name: GetLatestReadyDeliveryBySessionID :one
SELECT
  id,
  session_id,
  version,
  zip_key,
  zip_size_bytes,
  generated_at
FROM deliveries
WHERE session_id = sqlc.arg(session_id)
  AND status = 'ready'
ORDER BY version DESC
LIMIT 1;
