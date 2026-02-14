/* --- Users --- */

-- name: CreateUser :one
INSERT INTO users (id, username, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1 LIMIT 1 FOR UPDATE;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 LIMIT 1 FOR UPDATE;

-- name: UpdateUserPassword :one
UPDATE users 
SET password_hash = $2 
WHERE id = $1 
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;