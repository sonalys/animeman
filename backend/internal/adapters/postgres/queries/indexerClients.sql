-- name: CreateIndexerClient :one
INSERT INTO indexer_clients (
    id, owner_id, address, type, auth_id
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetIndexerClient :one
SELECT 
    c.id, c.owner_id, c.address, c.type,
    a.id as auth_id, a.type as auth_type, a.credentials as auth_credentials
FROM indexer_clients c
JOIN authentications a ON c.auth_id = a.id
WHERE c.id = $1 LIMIT 1 FOR UPDATE;

-- name: ListIndexerClientsByOwner :many
SELECT 
    c.id, c.owner_id, c.address, c.type,
    a.id as auth_id, a.type as auth_type, a.credentials as auth_credentials
FROM indexer_clients c
JOIN authentications a ON c.auth_id = a.id
WHERE c.owner_id = $1
ORDER BY c.id ASC;

-- name: UpdateIndexerAddress :exec
UPDATE indexer_clients
SET address = $2
WHERE id = $1;

-- name: DeleteIndexerClient :exec
-- Note: ON DELETE CASCADE handles the auth deletion if schema is set that way
DELETE FROM indexer_clients
WHERE id = $1;