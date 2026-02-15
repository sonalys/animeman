-- name: CreateWatchlist :one
INSERT INTO watchlists (
    id,
    owner_id, 
    source, 
    external_id, 
    sync_frequency,
    created_at
) VALUES (
    sqlc.arg(id),
    sqlc.arg(owner_id), 
    sqlc.arg(source)::watchlist_source, 
    sqlc.arg(external_id), 
    sqlc.arg(sync_frequency),
    sqlc.arg(created_at)
) RETURNING *;

-- name: GetWatchlistByID :one
SELECT * FROM watchlists
WHERE id = sqlc.arg(id) LIMIT 1;

-- name: ListWatchlistsByOwner :many
SELECT * FROM watchlists
WHERE owner_id = sqlc.arg(owner_id)
ORDER BY created_at DESC;

-- name: UpdateWatchlistSync :one
UPDATE watchlists
SET 
    last_synced_at = NOW(),
    sync_frequency = sqlc.arg(sync_frequency)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CreateWatchlistEntry :one
INSERT INTO watchlist_entries (
    watchlist_id, 
    media_id, 
    season_id, 
    last_watched_id, 
    status
) VALUES (
    sqlc.arg(watchlist_id), 
    sqlc.arg(media_id), 
    sqlc.arg(season_id), 
    sqlc.arg(last_watched_id), 
    sqlc.arg(status)::watchlist_status
)
RETURNING *;

-- name: GetWatchlistEntries :many
SELECT * FROM watchlist_entries
WHERE watchlist_id = sqlc.arg(watchlist_id)
ORDER BY updated_at DESC;

-- name: DeleteWatchlistEntry :exec
DELETE FROM watchlist_entries
WHERE id = sqlc.arg(id);

-- name: DeleteWatchlist :exec
DELETE FROM watchlists
WHERE id = sqlc.arg(id);