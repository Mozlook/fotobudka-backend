-- name: HasFinalPhotoForSessionPhoto :one
SELECT EXISTS (
  SELECT 1
  FROM final_photos
  WHERE session_id = sqlc.arg(session_id)
    AND photo_id = sqlc.arg(photo_id)
) AS ok;

-- name: InsertFinalPhoto :exec
INSERT INTO final_photos (
  id,
  session_id,
  photo_id,
  final_key,
  final_size_bytes,
  created_at
) VALUES (
  sqlc.arg(id),
  sqlc.arg(session_id),
  sqlc.arg(photo_id),
  sqlc.arg(final_key),
  sqlc.arg(final_size_bytes),
  now()
);
