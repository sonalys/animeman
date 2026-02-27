package monitoring

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/ports"
)

type (
	collectionMonitor struct {
		repository               ports.CollectionRepository
		shutdown                 func()
		wg                       sync.WaitGroup
		collectionWatchersResync map[collections.CollectionID]func()
		lock                     sync.Mutex
	}
)

func New(
	repository ports.CollectionRepository,
) *collectionMonitor {
	return &collectionMonitor{
		repository:               repository,
		collectionWatchersResync: make(map[collections.CollectionID]func()),
	}
}

func (m *collectionMonitor) Start(ctx context.Context) error {
	ctx, shutdown := context.WithCancel(ctx)
	defer shutdown()

	m.shutdown = shutdown

	if err := m.initExisting(ctx); err != nil {
		return err
	}

	for notification := range m.repository.Listen(ctx) {
		switch notification.Action {
		case ports.RepositoryActionCreate:
			collection, err := m.repository.Get(ctx, notification.ID)
			if err != nil {
				log.Error().
					Err(err).
					Stringer("collectionID", notification.ID).
					Msg("Could not read collection")
				continue
			}

			go m.startWatch(ctx, collection)
		case ports.RepositoryActionUpdate:
			collection, err := m.repository.Get(ctx, notification.ID)
			if err != nil {
				log.Error().
					Err(err).
					Stringer("collectionID", notification.ID).
					Msg("Could not get read collection")
				continue
			}

			if !collection.Monitored {
				m.stopWatch(notification.ID)
			}
		case ports.RepositoryActionDelete:
			m.stopWatch(notification.ID)
		default:
			log.Error().
				Msg("Received an unknown repository notification action")
		}
	}

	return nil
}

func (m *collectionMonitor) initExisting(ctx context.Context) error {
	existingCollections, err := m.repository.List(ctx, ports.ListOptions{PageSize: 10})
	if err != nil {
		return fmt.Errorf("listing existing collections: %w", err)
	}

	for _, collection := range existingCollections {
		go m.startWatch(ctx, &collection)
	}

	return nil
}

func (m *collectionMonitor) stopWatch(id collections.CollectionID) {
	m.lock.Lock()
	defer m.lock.Unlock()

	shutdown, exists := m.collectionWatchersResync[id]
	if !exists {
		return
	}

	shutdown()

	log.Debug().
		Stringer("collectionID", id).
		Msg("Triggered collection watch removal")

	delete(m.collectionWatchersResync, id)
}

func (m *collectionMonitor) startWatch(ctx context.Context, collection *collections.Collection) {
	m.lock.Lock()
	ctx, cancel := context.WithCancel(ctx)
	m.collectionWatchersResync[collection.ID] = cancel
	m.lock.Unlock()

	routine := newRoutine(collection)

	routine.start(ctx)
}
