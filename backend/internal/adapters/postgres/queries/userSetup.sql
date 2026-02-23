-- name: IsUserSetupComplete :one
SELECT is_completed FROM user_setup WHERE user_id = $1;

-- name: CompleteUserSetup :exec
UPDATE user_setup SET is_completed = true WHERE user_id = $1;