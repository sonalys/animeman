-- name: CreateTransferClient :one
INSERT INTO transfer_clients (
    id, owner_id, address, type, auth_id
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetTransferClient :one
-- Reconstructs the TransferClient aggregate with nested Authentication
SELECT 
    tc.id, tc.owner_id, tc.address, tc.type,
    a.id as auth_id, a.type as auth_type, a.credentials as auth_credentials
FROM transfer_clients tc
JOIN authentications a ON tc.auth_id = a.id
WHERE tc.id = $1 LIMIT 1 FOR UPDATE;

-- name: ListTransferClientsByOwner :many
SELECT 
    tc.id, tc.owner_id, tc.address, tc.type,
    a.id as auth_id, a.type as auth_type, a.credentials as auth_credentials
FROM transfer_clients tc
JOIN authentications a ON tc.auth_id = a.id
WHERE tc.owner_id = $1
ORDER BY tc.id;

-- name: UpdateTransferClientAddress :exec
UPDATE transfer_clients
SET address = $2
WHERE id = $1;

-- name: DeleteTransferClient :exec
-- Cascade will handle the associated row in authentications
DELETE FROM transfer_clients
WHERE id = $1;