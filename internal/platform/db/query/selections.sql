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

-- name: GetSessionSubmitDataForUpdate :one
SELECT
  id,
  status,
  base_price_cents,
  included_count,
  extra_price_cents,
  min_select_count,
  currency,
  payment_mode
FROM sessions
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: CountSelectionsBySessionID :one
SELECT COUNT(*)::bigint AS selected_count
FROM selections
WHERE session_id = sqlc.arg(session_id);

-- name: InsertSessionPayment :exec
INSERT INTO payments (
  id,
  session_id,
  method,
  status,
  amount_cents,
  created_at,
  updated_at
) VALUES (
  sqlc.arg(id),
  sqlc.arg(session_id),
  sqlc.arg(method),
  sqlc.arg(status),
  sqlc.arg(amount_cents),
  now(),
  now()
);

-- name: MarkSessionWaitingForPayment :execrows
UPDATE sessions
SET
  status = 'waiting_for_payment',
  updated_at = now()
WHERE id = sqlc.arg(id)
  AND status = 'selecting';

-- name: CountSelectedPhotosWithoutFinal :one
SELECT COUNT(*)::bigint
FROM selections s
LEFT JOIN final_photos fp
  ON fp.session_id = s.session_id
 AND fp.photo_id = s.photo_id
WHERE s.session_id = sqlc.arg(session_id)
  AND fp.id IS NULL;
