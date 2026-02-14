-- name: CreateSeason :one
INSERT INTO seasons (
    id, media_id, number, airing_status, metadata
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetSeason :one
SELECT * FROM seasons 
WHERE id = $1 LIMIT 1 FOR UPDATE;

-- name: GetSeasonByNumber :one
-- Useful for finding "Season 2" of a specific show
SELECT * FROM seasons 
WHERE media_id = $1 AND number = $2 LIMIT 1;

-- name: ListSeasonsByMedia :many
SELECT * FROM seasons 
WHERE media_id = $1 
ORDER BY number ASC;

-- name: UpdateSeasonMetadata :one
UPDATE seasons
SET 
    airing_status = $2,
    metadata = $3
WHERE id = $1
RETURNING *;

-- name: DeleteSeason :exec
DELETE FROM seasons WHERE id = $1;