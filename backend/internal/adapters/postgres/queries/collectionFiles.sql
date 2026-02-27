-- name: RegisterCollectionFile :one
INSERT INTO collection_files (
    id, episode_id, season_id, media_id, relative_path, 
    size_bytes, release_group, version, source,
    video_info, audio_streams, subtitle_streams, chapters, hashes,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
)
RETURNING *;

-- name: GetCollectionFile :one
SELECT * FROM collection_files 
WHERE id = $1 FOR UPDATE;

-- name: GetCollectionFileByEpisode :one
SELECT * FROM collection_files 
WHERE episode_id = $1 FOR UPDATE;

-- name: ListCollectionFilesBySeason :many
SELECT * FROM collection_files 
WHERE season_id = $1
ORDER BY relative_path ASC;

-- name: ListCollectionFilesPaginated :many
SELECT * FROM collection_files
WHERE 
    sqlc.narg(last_id)::uuid is NULL OR id < sqlc.narg(last_id)::uuid AND
    sqlc.narg(collection_id)::uuid is NULL OR collection_id = sqlc.narg(collection_id)::uuid
ORDER BY id DESC
LIMIT sqlc.narg('limit')::integer;

-- name: UpdateCollectionFile :one
UPDATE collection_files
SET 
    relative_path = $2, 
    size_bytes = $3, 
    version = $4,
    video_info = $5,
    audio_streams = $6,
    subtitle_streams = $7,
    chapters = $8,
    hashes = $9
WHERE id = $1
RETURNING *;

-- name: DeleteCollectionFile :exec
DELETE FROM collection_files WHERE id = $1;