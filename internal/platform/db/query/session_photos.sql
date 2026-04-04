-- name: GetSessionPhotoByIDAndSessionID :one
SELECT
    id,
    session_id,
    source_key,
    status,
    source_size_bytes,
    watermark_seed
FROM session_photos
WHERE id = $1
AND session_id = $2;

-- name: MarkSessionPhotoUploaded :execrows
UPDATE session_photos
SET
  status = 'uploaded',
  source_size_bytes = $1
WHERE id = $2
AND session_id = $3;

