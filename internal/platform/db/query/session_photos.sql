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

-- name: MarkPhotoProcessing :execrows
UPDATE session_photos
SET
    status = 'processing'
WHERE id = sqlc.arg(id)
  AND session_id = sqlc.arg(session_id)
  AND status IN ('uploaded', 'processing');

-- name: MarkPhotoReady :execrows
UPDATE session_photos
SET
status ='ready',
thumb_key = sqlc.arg(thumb_key),
proof_key = sqlc.arg(proof_key) 
WHERE id = sqlc.arg(id) 
AND session_id = sqlc.arg(session_id)
AND status = 'processing';

-- name: MarkPhotoFailed :execrows
UPDATE session_photos
SET
status ='failed'
WHERE id = sqlc.arg(id) 
AND session_id = sqlc.arg(session_id)
AND status IN ('uploaded', 'processing');

-- name: GetSessionPhotoStats :one
SELECT
    COUNT(*)::bigint AS total_count,
    COUNT(*) FILTER (WHERE status = 'pending_upload')::bigint AS pending_upload_count,
    COUNT(*) FILTER (WHERE status = 'uploaded')::bigint       AS uploaded_count,
    COUNT(*) FILTER (WHERE status = 'processing')::bigint     AS processing_count,
    COUNT(*) FILTER (WHERE status = 'ready')::bigint          AS ready_count,
    COUNT(*) FILTER (WHERE status = 'failed')::bigint         AS failed_count
FROM session_photos
WHERE session_id = sqlc.arg(session_id);
