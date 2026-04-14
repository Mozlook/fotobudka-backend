-- name: UpsertSelection :exec
INSERT INTO selections (
    session_id,
    photo_id,
    selected_at,
    note
) VALUES (
    sqlc.arg(session_id),
    sqlc.arg(photo_id),
    now(),
    sqlc.narg(note)
)
ON CONFLICT (session_id, photo_id)
DO UPDATE
SET note = EXCLUDED.note;

-- name: DeleteSelection :execrows
DELETE FROM selections
WHERE session_id = sqlc.arg(session_id)
  AND photo_id = sqlc.arg(photo_id);
