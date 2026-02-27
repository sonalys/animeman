package monitoring

import (
	"context"
	"io/fs"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/utils/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type routine struct {
	collection *collections.Collection
}

func newRoutine(collection *collections.Collection) *routine {
	return &routine{
		collection: collection,
	}
}

func (r *routine) start(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().
				Ctx(ctx).
				Any("r", r).
				Msg("Recovered from panic")
		}
	}()

	logger := log.With().
		Ctx(ctx).
		Stringer("collectionID", r.collection.ID).
		Logger()

	logger.Info().
		Msg("Started collection watch routine")

	defer logger.Info().
		Msg("Stopped collection watch routine")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()

	err = filepath.WalkDir(r.collection.BasePath, func(path string, entry fs.DirEntry, err error) error {
		return watcher.Add(path)
	})
	if err != nil {
		logger.
			Err(err).
			Msg("Error recursively watching collection directory")

		return
	}

	for {
		select {
		case <-ctx.Done():

			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			r.handleEvent(ctx, event)
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			logger.Error().
				Err(err).
				Msg("Received error from filesystem watcher")
		}
	}
}

func (r *routine) handleEvent(ctx context.Context, event fsnotify.Event) {
	ctx, span := otel.Tracer.Start(ctx, "collectionMonitor.handleEvent",
		trace.WithAttributes(
			attribute.Stringer("collectionID", r.collection.ID),
		),
	)
	defer span.End()

	logger := log.Ctx(ctx).With().
		Ctx(ctx).
		Logger()

	logger.Debug().
		Any("event", event).
		Msg("Received event")
}
