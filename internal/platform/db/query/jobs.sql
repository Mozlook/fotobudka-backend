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

-- name: ClaimDueJobs :many
WITH picked AS (
    SELECT id
    FROM jobs
    WHERE status = 'pending'
      AND next_run_at <= now()
      AND attempts < max_attempts
    ORDER BY next_run_at, created_at
    LIMIT sqlc.arg(limit_count)
    FOR UPDATE SKIP LOCKED
)
UPDATE jobs
SET
    status = 'running',
    locked_at = now(),
    locked_by = sqlc.arg(locked_by),
    attempts = attempts + 1,
    updated_at = now(),
    last_error = NULL
WHERE id IN (SELECT id FROM picked)
RETURNING
    id,
    type,
    status,
    payload,
    attempts,
    max_attempts,
    next_run_at,
    locked_at,
    locked_by,
    last_error,
    created_at,
    updated_at;

-- name: MarkJobSucceeded :execrows
UPDATE jobs
SET
    status = 'succeeded',
    locked_at = NULL,
    locked_by = NULL,
    last_error = NULL,
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND status = 'running';

-- name: MarkJobRetry :execrows
UPDATE jobs
SET
    status = 'pending',
    locked_at = NULL,
    locked_by = NULL,
    last_error = sqlc.arg(last_error),
    next_run_at = sqlc.arg(next_run_at),
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND status = 'running';

-- name: MarkJobFailed :execrows
UPDATE jobs
SET
    status = 'failed',
    locked_at = NULL,
    locked_by = NULL,
    last_error = sqlc.arg(last_error),
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND status = 'running';
