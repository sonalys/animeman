-- name: CreateUser :exec
INSERT INTO users (id,username,password_hash) 
VALUES ($1,$2,$3);

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1 FOR UPDATE;

-- name: UpdateUser :exec
UPDATE users 
SET 
    id = $1,
    username = $2,
    password_hash = $3
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;