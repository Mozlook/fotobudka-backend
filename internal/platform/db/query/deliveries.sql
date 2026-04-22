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
