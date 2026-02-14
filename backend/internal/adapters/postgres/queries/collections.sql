-- name: CreateCollection :one
INSERT INTO collections (
    id, owner_id, name, base_path, tags, monitored, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetCollection :one
SELECT * FROM collections
WHERE id = $1 LIMIT 1 FOR UPDATE;

-- name: ListCollectionsByOwner :many
SELECT * FROM collections
WHERE owner_id = $1
ORDER BY created_at DESC;

-- name: FindCollectionsByTag :many
-- Uses the GIN index to find collections containing the specified tag
SELECT * FROM collections
WHERE $1 = ANY(tags) AND owner_id = $2;

-- name: UpdateCollection :one
UPDATE collections
SET 
    name = $2,
    base_path = $3,
    tags = $4,
    monitored = $5
WHERE id = $1
RETURNING *;

-- name: SetMonitoredStatus :exec
-- Useful for bulk actions in the UI
UPDATE collections
SET monitored = $2
WHERE id = $1;

-- name: UpdateCollectionTags :exec
UPDATE collections
SET tags = $2
WHERE id = $1;

-- name: DeleteCollection :exec
DELETE FROM collections
WHERE id = $1;