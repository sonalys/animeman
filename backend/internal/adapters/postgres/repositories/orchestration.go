package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/mappers"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/orchestration"
	"github.com/sonalys/animeman/internal/ports"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
)

type taskRepository struct {
	conn *pgxpool.Pool
}

func NewTaskRepository(conn *pgxpool.Pool) ports.TaskRepository {
	return &taskRepository{
		conn: conn,
	}
}

func (r *taskRepository) CreateTask(ctx context.Context, t *orchestration.Task) error {
	queries := sqlcgen.New(r.conn)
	params := sqlcgen.CreateTaskParams{
		ID:         t.ID,
		TaskType:   t.Type,
		Payload:    t.Payload,
		MaxRetries: int32(t.MaxRetries),
		TraceID: pgtype.Text{
			String: t.TraceID,
			Valid:  true,
		},
		SpanID: pgtype.Text{
			String: t.SpanID,
			Valid:  true,
		},
	}
	if _, err := queries.CreateTask(ctx, params); err != nil {
		return handleWriteError(err, nil)
	}
	return nil
}

func (r *taskRepository) ClaimTask(ctx context.Context, timeout time.Duration) (*orchestration.Task, error) {
	queries := sqlcgen.New(r.conn)

	model, err := queries.ClaimNextTask(ctx, pgtype.Interval{
		Microseconds: timeout.Microseconds(),
	})
	if err != nil {
		return nil, handleReadError(err)
	}

	return new(mappers.NewTask(model)), nil
}

func (r *taskRepository) MarkCompleted(ctx context.Context, id orchestration.TaskID) error {
	queries := sqlcgen.New(r.conn)
	if err := queries.CompleteTask(ctx, id); err != nil {
		return handleWriteError(err, nil)
	}
	return nil
}

func (r *taskRepository) MarkFailed(ctx context.Context, id orchestration.TaskID, nextRetry time.Time) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.FailTaskParams{
		ID: id,
		NextRetryAt: pgtype.Timestamptz{
			Time:  nextRetry,
			Valid: true,
		},
	}
	if err := queries.FailTask(ctx, params); err != nil {
		return handleWriteError(err, nil)
	}
	return nil
}

func (r *taskRepository) AddLog(ctx context.Context, entry *orchestration.TaskLog) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.AddTaskLogParams{
		TaskID:  entry.TaskID,
		Level:   mappers.NewLogLevelModel(entry.Level),
		Message: entry.Message,
		TraceID: pgtype.Text{
			String: entry.TraceID,
			Valid:  true,
		},
		SpanID: pgtype.Text{
			String: entry.SpanID,
			Valid:  true,
		},
	}

	if err := queries.AddTaskLog(ctx, params); err != nil {
		return handleWriteError(err, nil)
	}

	return nil
}

func (r *taskRepository) ListTasks(ctx context.Context, lastID orchestration.TaskID, pageSize int32) ([]orchestration.Task, error) {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.ListTasksPaginatedParams{
		LastID:   lastID,
		PageSize: pageSize,
	}

	models, err := queries.ListTasksPaginated(ctx, params)
	if err != nil {
		return nil, handleReadError(err)
	}

	return sliceutils.Map(models, mappers.NewTask), nil
}

func (r *taskRepository) ListTaskLogs(ctx context.Context, taskID orchestration.TaskID, lastID orchestration.TaskLogID, pageSize int32) ([]orchestration.TaskLog, error) {
	queries := sqlcgen.New(r.conn)
	params := sqlcgen.ListTaskLogsPaginatedParams{
		TaskID:   taskID,
		LastID:   lastID,
		PageSize: pageSize,
	}

	models, err := queries.ListTaskLogsPaginated(ctx, params)
	if err != nil {
		return nil, handleReadError(err)
	}

	return sliceutils.Map(models, mappers.NewTaskLog), nil
}

func (r *taskRepository) RotateLogs(ctx context.Context, retention time.Duration, maxLogs int32) error {
	queries := sqlcgen.New(r.conn)

	return queries.RotateLogs(ctx, sqlcgen.RotateLogsParams{
		RetentionInterval: pgtype.Interval{Microseconds: retention.Microseconds()},
		MaxLogs:           maxLogs,
	})
}
