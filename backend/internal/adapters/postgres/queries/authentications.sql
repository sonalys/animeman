-- name: CreateAuthentication :one
INSERT INTO authentications (
    id, type, credentials
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetAuthentication :one
SELECT * FROM authentications
WHERE id = $1 LIMIT 1 FOR UPDATE;

-- name: UpdateCredentials :exec
UPDATE authentications
SET credentials = $2
WHERE id = $1;

-- name: DeleteAuthentication :exec
DELETE FROM authentications
WHERE id = $1;

-- name: ListAuthenticationsByType :many
SELECT * FROM authentications
WHERE type = $1
ORDER BY id;