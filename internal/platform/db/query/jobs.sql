-- name: EnqueueJob :exec
INSERT INTO jobs (
    id,
    type,
    payload,
    max_attempts
) VALUES (
    sqlc.arg(id),
    sqlc.arg(type),
    sqlc.arg(payload),
    sqlc.arg(max_attempts)
);
