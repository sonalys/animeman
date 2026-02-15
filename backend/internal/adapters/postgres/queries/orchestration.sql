-- name: CreateTask :one
INSERT INTO orchestration_tasks (
    id, task_type, status, payload, retry_count, max_retries, 
    next_retry_at, trace_id, span_id, created_at, updated_at
) VALUES (
    @id, @task_type, @status, @payload, @retry_count, @max_retries, 
    @next_retry_at, @trace_id, @span_id, @created_at, @updated_at
) RETURNING *;

-- name: ClaimNextTask :one
-- Claims a task if it is pending/retrying OR if it has expired (crashed)
UPDATE orchestration_tasks
SET 
    status = 'running',
    expires_at = NOW() + @timeout_interval::interval,
    updated_at = NOW()
WHERE id = (
    SELECT id 
    FROM orchestration_tasks 
    WHERE (
        (status = 'pending' AND (next_retry_at IS NULL OR next_retry_at <= NOW()))
        OR 
        (status = 'running' AND expires_at <= NOW())
    )
    ORDER BY created_at ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING *;

-- name: CompleteTask :exec
UPDATE orchestration_tasks
SET status = 'completed', expires_at = NULL, updated_at = NOW()
WHERE id = @id;

-- name: FailTask :exec
UPDATE orchestration_tasks
SET 
    status = CASE WHEN retry_count < max_retries THEN 'pending'::text ELSE 'failed'::text END,
    retry_count = retry_count + 1,
    next_retry_at = @next_retry_at,
    expires_at = NULL,
    updated_at = NOW()
WHERE id = @id;

-- name: AddTaskLog :exec
INSERT INTO task_logs (
    id, task_id, level, message, trace_id, span_id, created_at
) VALUES (
    @id, @task_id, @level::log_level, @message, @trace_id, @span_id, @created_at
);

-- name: RotateLogs :exec
-- Optimized for your idx_task_logs_rotation index.
-- Deletes logs based on time AND a hard count limit.
DELETE FROM task_logs
WHERE created_at < (NOW() - sqlc.arg(retention_interval)::interval)
   OR id NOT IN (
       SELECT id FROM task_logs
       ORDER BY created_at DESC
       LIMIT sqlc.arg(max_logs)
   );

-- name: ListTasksPaginated :many
-- Uses keyset pagination. Provide a zero UUID for the first page.
SELECT * FROM orchestration_tasks
WHERE id > @last_id
ORDER BY id ASC
LIMIT @page_size;

-- name: ListTasksByStatusPaginated :many
SELECT * FROM orchestration_tasks
WHERE status = @status 
  AND id > @last_id
ORDER BY id ASC
LIMIT @page_size;

-- name: ListTaskLogsPaginated :many
-- Provides a list of logs for a specific task with pagination.
SELECT * FROM task_logs
WHERE task_id = @task_id
  AND id > @last_id
ORDER BY id ASC
LIMIT @page_size;

-- name: ListAllLogsPaginated :many
-- Global log feed for the dashboard.
SELECT * FROM task_logs
WHERE id > @last_id
ORDER BY id ASC
LIMIT @page_size;