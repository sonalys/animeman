-- name: CreateEpisode :one
INSERT INTO episodes (
    id, season_id, media_id, type, number, titles, airing_date
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetEpisode :one
SELECT * FROM episodes 
WHERE id = $1 LIMIT 1 FOR UPDATE;

-- name: GetEpisodeByNumber :one
-- Useful for identifying a specific episode file during a disk scan
SELECT * FROM episodes 
WHERE season_id = $1 AND number = $2 LIMIT 1;

-- name: ListEpisodesBySeason :many
SELECT * FROM episodes 
WHERE season_id = $1 
ORDER BY airing_date ASC;

-- name: UpdateEpisodeMetadata :one
UPDATE episodes
SET 
    type = $2,
    titles = $3,
    airing_date = $4
WHERE id = $1
RETURNING *;

-- name: DeleteEpisode :exec
DELETE FROM episodes WHERE id = $1;