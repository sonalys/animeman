/* --- Users --- */

-- name: CreateUser :one
INSERT INTO users (id, username, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 LIMIT 1;

-- name: UpdateUserPassword :one
UPDATE users 
SET password_hash = $2 
WHERE id = $1 
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;


/* --- Anime Lists --- */

-- name: CreateAnimeList :one
INSERT INTO anime_lists (id, owner_id, remote_username, source)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAnimeList :one
-- Lookup by the unique composite primary key
SELECT * FROM anime_lists 
WHERE source = $1 AND remote_username = $2 LIMIT 1;

-- name: ListAnimeListsByOwner :many
SELECT * FROM anime_lists WHERE owner_id = $1;

-- name: UpdateAnimeListOwner :one
UPDATE anime_lists 
SET owner_id = $3
WHERE source = $1 AND remote_username = $2
RETURNING *;

-- name: DeleteAnimeList :exec
DELETE FROM anime_lists WHERE source = $1 AND remote_username = $2;


/* --- Torrent Clients --- */

-- name: CreateTorrentClient :one
INSERT INTO torrent_clients (id, owner_id, source, host, username, password)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetTorrentClientByHost :one
SELECT * FROM torrent_clients WHERE host = $1 LIMIT 1;

-- name: ListTorrentClientsByOwner :many
SELECT * FROM torrent_clients WHERE owner_id = $1;

-- name: UpdateTorrentClientCredentials :one
UPDATE torrent_clients
SET username = $2, password = $3
WHERE host = $1
RETURNING *;

-- name: DeleteTorrentClient :exec
DELETE FROM torrent_clients WHERE host = $1;


/* --- Prowlarr Configurations --- */

-- name: CreateProwlarrConfiguration :one
INSERT INTO prowlarr_configurations (id, owner_id, host, api_key)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProwlarrConfigurationByOwner :one
SELECT * FROM prowlarr_configurations WHERE owner_id = $1 LIMIT 1 FOR UPDATE;

-- name: UpdateProwlarrConfiguration :one
UPDATE prowlarr_configurations
SET 
    api_key = $2,
    host = $3
WHERE id = $1
RETURNING *;

-- name: DeleteProwlarrConfiguration :exec
DELETE FROM prowlarr_configurations WHERE id = $1;


/* --- Import Configurations --- */

-- name: CreateImportConfiguration :one
INSERT INTO import_configurations (
    id, owner_id, anime_list_id, torrent_client_id, prowlarr_configuration_id
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetImportConfiguration :one
-- Lookup by unique composite tuple
SELECT * FROM import_configurations 
WHERE anime_list_id = $1 AND torrent_client_id = $2 LIMIT 1;

-- name: UpdateImportProwlarrConfig :one
UPDATE import_configurations
SET prowlarr_configuration_id = $3
WHERE anime_list_id = $1 AND torrent_client_id = $2
RETURNING *;

-- name: DeleteImportConfiguration :exec
DELETE FROM import_configurations 
WHERE anime_list_id = $1 AND torrent_client_id = $2;