-- name: GetPublicPhotographerByUsername :one
SELECT
    user_id,
    username,
    display_name,
    bio,
    social_links,
    created_at,
    updated_at
FROM photographer_profiles
WHERE username = $1;

-- name: ListPublicGalleriesByPhotographerID :many
SELECT
    g.id,
    g.photographer_id,
    g.title,
    g.slug,
    g.is_public,
    g.created_at,
    COALESCE(stats.photo_count, 0)::int AS photo_count,
    COALESCE(cover.image_key, '')::text AS cover_image_key
FROM galleries g
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS photo_count
    FROM gallery_photos gp
    WHERE gp.gallery_id = g.id
      AND gp.width > 0
      AND gp.height > 0
) stats ON true
LEFT JOIN LATERAL (
    SELECT gp.image_key
    FROM gallery_photos gp
    WHERE gp.gallery_id = g.id
      AND gp.width > 0
      AND gp.height > 0
    ORDER BY gp.sort_order ASC, gp.created_at ASC
    LIMIT 1
) cover ON true
WHERE g.photographer_id = $1
  AND g.is_public = true
ORDER BY g.created_at DESC;

-- name: GetPublicGalleryByUsernameAndSlug :one
SELECT
    g.id,
    g.photographer_id,
    g.title,
    g.slug,
    g.is_public,
    g.created_at
FROM galleries g
JOIN photographer_profiles pp
    ON pp.user_id = g.photographer_id
WHERE pp.username = $1
  AND g.slug = $2
  AND g.is_public = true;

-- name: ListPublicGalleryPhotos :many
SELECT
    id,
    gallery_id,
    image_key,
    width,
    height,
    sort_order,
    created_at
FROM gallery_photos
WHERE gallery_id = $1
  AND width > 0
  AND height > 0
ORDER BY sort_order ASC, created_at ASC;

-- name: ListGalleriesByOwner :many
SELECT
    g.id,
    g.photographer_id,
    g.title,
    g.slug,
    g.is_public,
    g.created_at,
    COALESCE(stats.photo_count, 0)::int AS photo_count,
    COALESCE(cover.image_key, '')::text AS cover_image_key
FROM galleries g
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS photo_count
    FROM gallery_photos gp
    WHERE gp.gallery_id = g.id
      AND gp.width > 0
      AND gp.height > 0
) stats ON true
LEFT JOIN LATERAL (
    SELECT gp.image_key
    FROM gallery_photos gp
    WHERE gp.gallery_id = g.id
      AND gp.width > 0
      AND gp.height > 0
    ORDER BY gp.sort_order ASC, gp.created_at ASC
    LIMIT 1
) cover ON true
WHERE g.photographer_id = $1
ORDER BY g.created_at DESC;

-- name: GetGalleryByIDForOwner :one
SELECT
    g.id,
    g.photographer_id,
    g.title,
    g.slug,
    g.is_public,
    g.created_at,
    COALESCE(stats.photo_count, 0)::int AS photo_count,
    COALESCE(cover.image_key, '')::text AS cover_image_key
FROM galleries g
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS photo_count
    FROM gallery_photos gp
    WHERE gp.gallery_id = g.id
      AND gp.width > 0
      AND gp.height > 0
) stats ON true
LEFT JOIN LATERAL (
    SELECT gp.image_key
    FROM gallery_photos gp
    WHERE gp.gallery_id = g.id
      AND gp.width > 0
      AND gp.height > 0
    ORDER BY gp.sort_order ASC, gp.created_at ASC
    LIMIT 1
) cover ON true
WHERE g.id = $1
  AND g.photographer_id = $2;

-- name: CreateGallery :one
INSERT INTO galleries (
    id,
    photographer_id,
    title,
    slug,
    is_public
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING
    id,
    photographer_id,
    title,
    slug,
    is_public,
    created_at;

-- name: UpdateGallery :one
UPDATE galleries
SET
    title = $3,
    slug = $4,
    is_public = $5
WHERE id = $1
  AND photographer_id = $2
RETURNING
    id,
    photographer_id,
    title,
    slug,
    is_public,
    created_at;

-- name: DeleteGallery :exec
DELETE FROM galleries
WHERE id = $1
  AND photographer_id = $2;

-- name: ListGalleryPhotosForOwner :many
SELECT
    gp.id,
    gp.gallery_id,
    gp.image_key,
    gp.width,
    gp.height,
    gp.sort_order,
    gp.created_at
FROM gallery_photos gp
JOIN galleries g
    ON g.id = gp.gallery_id
WHERE gp.gallery_id = $1
  AND g.photographer_id = $2
ORDER BY gp.sort_order ASC, gp.created_at ASC;

-- name: GetGalleryPhotoForOwner :one
SELECT
    gp.id,
    gp.gallery_id,
    gp.image_key,
    gp.width,
    gp.height,
    gp.sort_order,
    gp.created_at
FROM gallery_photos gp
JOIN galleries g
    ON g.id = gp.gallery_id
WHERE gp.id = $1
  AND gp.gallery_id = $2
  AND g.photographer_id = $3;


-- name: CreateGalleryPhotoFromCompletedUpload :one
INSERT INTO gallery_photos (
    id,
    gallery_id,
    image_key,
    width,
    height,
    sort_order
)
SELECT
    sqlc.arg(id),
    sqlc.arg(gallery_id),
    sqlc.arg(image_key),
    sqlc.arg(width),
    sqlc.arg(height),
    COALESCE(MAX(sort_order) + 1, 0)
FROM gallery_photos
WHERE gallery_id = sqlc.arg(gallery_id)
RETURNING
    id,
    gallery_id,
    image_key,
    width,
    height,
    sort_order,
    created_at;

-- name: DeleteGalleryPhoto :exec
DELETE FROM gallery_photos gp
USING galleries g
WHERE gp.gallery_id = g.id
  AND gp.id = $1
  AND gp.gallery_id = $2
  AND g.photographer_id = $3;
