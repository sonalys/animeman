
-- name: CreateMedia :one
INSERT INTO media (
    id, collection_id, quality_profile_id, titles, 
    monitoring_status, monitored_since, genres, 
    airing_started_at, airing_ended_at, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetMedia :one
SELECT * FROM media WHERE id = $1 LIMIT 1 FOR UPDATE;

-- name: ListMediaPaginated :many
SELECT * FROM media
WHERE (id) < ($1::uuid)
ORDER BY id DESC
LIMIT $2;

-- name: SearchMediaByTitlePaginated :many
-- Paginates fuzzy search results based on the last seen score and ID
SELECT *, similarity(titles_search_vector, $1) as score
FROM media
WHERE titles_search_vector % $1
AND (similarity(titles_search_vector, $1), id) < ($2::float4, $3::uuid)
ORDER BY score DESC, id ASC
LIMIT $4;

-- name: FindMediaByExactTitle :many
-- Uses the GIN index on the titles JSONB for an exact match within the array
SELECT * FROM media
WHERE titles @> jsonb_build_array(jsonb_build_object('Value', $1::text));

-- name: UpdateMedia :exec
UPDATE media
SET 
    titles = $2,
    monitoring_status = $3, 
    monitored_since = $4,
    genres = $5,
    airing_started_at = $6,
    airing_ended_at = $7,
    quality_profile_id = $8
WHERE id = $1;

-- name: UpdateMediaQualityProfile :exec
UPDATE media
SET quality_profile_id = $2
WHERE id = $1;

-- name: DeleteMedia :exec
DELETE FROM media WHERE id = $1;