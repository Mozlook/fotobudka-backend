-- name: GetPhotographerSelectionOverview :one
SELECT
    s.id AS session_id,
    s.status AS session_status,
    COUNT(sel.photo_id)::int AS selected_count,

    COALESCE(p.status, '') AS payment_status,
    COALESCE(p.amount_cents, 0)::int AS payment_amount_cents,
    p.paid_at AS payment_paid_at
FROM sessions s
LEFT JOIN selections sel
    ON sel.session_id = s.id
LEFT JOIN LATERAL (
    SELECT
        payments.status,
        payments.amount_cents,
        payments.paid_at
    FROM payments
    WHERE payments.session_id = s.id
    ORDER BY payments.created_at DESC
    LIMIT 1
) p ON true
WHERE s.id = $1
GROUP BY
    s.id,
    s.status,
    p.status,
    p.amount_cents,
    p.paid_at;

-- name: ListPhotographerSelectedPhotos :many
SELECT
    sp.id AS photo_id,
    sp.original_filename,
    sp.thumb_key,
    COALESCE(sel.note, '') AS note,
    sel.selected_at,

    fp.final_id,
    COALESCE(fp.final_uploaded, false)::bool AS final_uploaded
FROM selections sel
JOIN session_photos sp
    ON sp.id = sel.photo_id
    AND sp.session_id = sel.session_id
LEFT JOIN LATERAL (
    SELECT
        final_photos.id AS final_id,
        (
            final_photos.final_key IS NOT NULL
            AND COALESCE(final_photos.final_size_bytes, 0) > 0
        ) AS final_uploaded
    FROM final_photos
    WHERE final_photos.session_id = sel.session_id
      AND final_photos.photo_id = sel.photo_id
    ORDER BY final_photos.created_at DESC
    LIMIT 1
) fp ON true
WHERE sel.session_id = $1
ORDER BY sel.selected_at ASC, sp.created_at ASC;
