-- name: CreateQualityProfile :one
INSERT INTO quality_profiles (
    id, name, min_resolution, max_resolution, 
    codec_preference, release_group_preference
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetQualityProfile :one
SELECT * FROM quality_profiles
WHERE id = $1 LIMIT 1 FOR UPDATE;

-- name: ListQualityProfiles :many
SELECT * FROM quality_profiles
ORDER BY name ASC;

-- name: UpdateQualityProfile :one
UPDATE quality_profiles
SET 
    name = $2,
    min_resolution = $3,
    max_resolution = $4,
    codec_preference = $5,
    release_group_preference = $6
WHERE id = $1
RETURNING *;

-- name: DeleteQualityProfile :exec
DELETE FROM quality_profiles
WHERE id = $1;