package monitoring

import (
	"context"
	"io/fs"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/ports"
)

type (
	collectionMonitor struct {
		repository ports.CollectionRepository
		shutdown   func()
		wg         sync.WaitGroup
		config     Config
	}

	Config struct {
		MaxWorkers     int
		PollInterval   time.Duration
		DefaultTimeout time.Duration
	}
)

func New() *collectionMonitor {
	return &collectionMonitor{}
}

func (m *collectionMonitor) Start(ctx context.Context) {
	ctx, shutdown := context.WithCancel(ctx)
	m.shutdown = shutdown

	for i := 0; i < m.config.MaxWorkers; i++ {
		go m.workerLoop(ctx)
	}
}

func (m *collectionMonitor) workerLoop(ctx context.Context) {
	ticker := time.NewTicker(max(m.config.PollInterval, time.Minute))
	defer ticker.Stop()

	m.wg.Add(1)
	defer m.wg.Done()

	for {
		m.routine(ctx)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (m *collectionMonitor) routine(ctx context.Context) {
	collection := collections.Collection{
		BasePath: "./",
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()

	err = filepath.WalkDir(collection.BasePath, func(path string, entry fs.DirEntry, err error) error {
		return watcher.Add(path)
	})
	if err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-watcher.Events:
			if !ok {
				return
			}

		case _, ok := <-watcher.Errors:
			if !ok {
				return
			}
		}
	}
}
