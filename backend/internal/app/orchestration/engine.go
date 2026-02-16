package orchestration

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/domain/orchestration"
	"github.com/sonalys/animeman/internal/ports"
)

type TaskProcessor func(ctx context.Context, task *orchestration.Task) error

type Engine struct {
	store      ports.TaskRepository
	processors map[string]TaskProcessor
	config     EngineConfig
}

type EngineConfig struct {
	MaxWorkers     int
	PollInterval   time.Duration
	DefaultTimeout time.Duration
}

func NewEngine(store ports.TaskRepository, cfg EngineConfig) *Engine {
	return &Engine{
		store:      store,
		processors: make(map[string]TaskProcessor),
		config:     cfg,
	}
}

func (e *Engine) Register(taskType string, fn TaskProcessor) {
	e.processors[taskType] = fn
}

func (e *Engine) Start(ctx context.Context) {
	for i := 0; i < e.config.MaxWorkers; i++ {
		go e.workerLoop(ctx)
	}

	go e.janitorLoop(ctx)
}

func (e *Engine) workerLoop(ctx context.Context) {
	ticker := time.NewTicker(e.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			task, err := e.store.ClaimTask(ctx, e.config.DefaultTimeout)
			if err != nil {
				continue
			}

			e.execute(ctx, task)
		}
	}
}

func (e *Engine) execute(ctx context.Context, task *orchestration.Task) {
	ctx = injectTelemetry(ctx, task.TraceID, task.SpanID)
	ctx = log.Ctx(ctx).
		Hook(newLoggerOrchestrationHook(e.store, task)).
		WithContext(ctx)

	log.Info().
		Ctx(ctx).
		Msg("Starting task execution")

	fn, ok := e.processors[task.Type]
	if !ok {
		log.Error().
			Ctx(ctx).
			Str("taskType", task.Type).
			Msg("Processor for task not found")
		return
	}

	if err := fn(ctx, task); err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Msg("Task execution failed")

		nextRetry, shouldRetry := task.CalculateBackoff(30 * time.Second)
		if !shouldRetry {
			log.Warn().
				Ctx(ctx).
				Err(err).
				Msg("Task ran out of retries")
		}

		if err := e.store.MarkFailed(ctx, task.ID, nextRetry); err != nil {
			log.Warn().
				Ctx(ctx).
				Err(err).
				Msg("Failed to mark task as failed. It will be retried when expired")
		}

		log.Info().
			Ctx(ctx).
			Time("nextRetry", nextRetry).
			Msg("Task scheduled for retry")
		return
	}

	if err := e.store.MarkCompleted(ctx, task.ID); err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Msg("Failed to mark task as failed. It will be retried when expired")
	}

	log.Info().
		Ctx(ctx).
		Msg("Task completed")
}

func (e *Engine) janitorLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := e.store.RotateLogs(ctx, 30*24*time.Hour, 500_000)
			if err != nil {
				log.Error().
					Err(err).
					Msg("failed to run janitor for task logs")
			}
		}
	}
}
