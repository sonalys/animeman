package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/mappers"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/domain/watchlists"
	"github.com/sonalys/animeman/internal/ports"
)

type watchlistRepository struct {
	conn *pgxpool.Pool
}

func NewWatchlistRepository(conn *pgxpool.Pool) ports.WatchlistRepository {
	return &watchlistRepository{
		conn: conn,
	}
}

func (r *watchlistRepository) Create(ctx context.Context, watchlist *watchlists.Watchlist) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateWatchlistParams{
		ID:         watchlist.ID,
		OwnerID:    watchlist.Owner,
		ExternalID: pgtype.Text{String: watchlist.ExternalID},
		SyncFrequency: pgtype.Interval{
			Months:       0,
			Days:         0,
			Microseconds: watchlist.SyncFrequency.Microseconds(),
			Valid:        true,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:  watchlist.CreatedAt,
			Valid: true,
		},
		Source: mappers.NewWatchlistSourceModel(watchlist.Source),
	}

	if _, err := queries.CreateWatchlist(ctx, params); err != nil {
		return handleWriteError(err)
	}

	return nil
}

func (r *watchlistRepository) CreateEntry(ctx context.Context, entry *watchlists.WatchlistEntry) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateWatchlistEntryParams{
		WatchlistID:   entry.WatchlistID,
		MediaID:       entry.MediaID,
		SeasonID:      entry.SeasonID,
		LastWatchedID: entry.LastWatched,
		Status:        mappers.NewWatchlistStatusModel(entry.Status),
	}

	if _, err := queries.CreateWatchlistEntry(ctx, params); err != nil {
		return handleWriteError(err)
	}

	return nil
}

func (r *watchlistRepository) Delete(ctx context.Context, id watchlists.WatchlistID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteWatchlist(ctx, id); err != nil {
		return handleReadError(err)
	}

	return nil
}

func (r *watchlistRepository) DeleteEntry(ctx context.Context, id watchlists.WatchlistEntryID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteWatchlistEntry(ctx, id); err != nil {
		return handleReadError(err)
	}

	return nil
}

func (r *watchlistRepository) List(ctx context.Context) ([]watchlists.Watchlist, error) {
	queries := sqlcgen.New(r.conn)

	models, err := queries.ListWatchlists(ctx)
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]watchlists.Watchlist, 0, len(models))

	for i := range models {
		model := &models[i]

		response = append(response, watchlists.Watchlist{
			ID:            model.ID,
			Owner:         model.OwnerID,
			Source:        mappers.NewWatchlistSource(model.Source),
			ExternalID:    model.ExternalID.String,
			LastSyncedAt:  model.LastSyncedAt.Time,
			CreatedAt:     model.CreatedAt.Time,
			SyncFrequency: time.Duration(model.SyncFrequency.Microseconds * int64(time.Microsecond)),
		})
	}

	return response, nil
}

func (r *watchlistRepository) ListByOwner(ctx context.Context, id shared.UserID) ([]watchlists.Watchlist, error) {
	queries := sqlcgen.New(r.conn)

	models, err := queries.ListWatchlistsByOwner(ctx, id)
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]watchlists.Watchlist, 0, len(models))

	for i := range models {
		model := &models[i]

		response = append(response, watchlists.Watchlist{
			ID:            model.ID,
			Owner:         model.OwnerID,
			Source:        mappers.NewWatchlistSource(model.Source),
			ExternalID:    model.ExternalID.String,
			LastSyncedAt:  model.LastSyncedAt.Time,
			CreatedAt:     model.CreatedAt.Time,
			SyncFrequency: time.Duration(model.SyncFrequency.Microseconds * int64(time.Microsecond)),
		})
	}

	return response, nil
}

func (r *watchlistRepository) ListEntries(ctx context.Context, id watchlists.WatchlistID) ([]watchlists.WatchlistEntry, error) {
	queries := sqlcgen.New(r.conn)

	models, err := queries.GetWatchlistEntries(ctx, id)
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]watchlists.WatchlistEntry, 0, len(models))

	for i := range models {
		model := &models[i]

		response = append(response, watchlists.WatchlistEntry{
			ID:          model.ID,
			WatchlistID: model.WatchlistID,
			MediaID:     model.MediaID,
			SeasonID:    model.SeasonID,
			LastWatched: model.LastWatchedID,
			Status:      mappers.NewWatchlistStatus(model.Status),
			CreatedAt:   model.CreatedAt.Time,
			UpdatedAt:   model.UpdatedAt.Time,
		})
	}

	return response, nil
}

func (r *watchlistRepository) Update(ctx context.Context, id watchlists.WatchlistID, updateHandler ports.UpdateHandler[watchlists.Watchlist]) error {
	return transaction(ctx, r.conn, func(queries *sqlcgen.Queries) error {
		model, err := queries.GetWatchlistByID(ctx, id)
		if err != nil {
			return handleReadError(err)
		}

		watchlist := &watchlists.Watchlist{
			ID:            model.ID,
			Owner:         model.OwnerID,
			Source:        mappers.NewWatchlistSource(model.Source),
			ExternalID:    model.ExternalID.String,
			CreatedAt:     model.CreatedAt.Time,
			LastSyncedAt:  model.LastSyncedAt.Time,
			SyncFrequency: time.Duration(model.SyncFrequency.Microseconds * int64(time.Microsecond)),
		}

		if err := updateHandler(watchlist); err != nil {
			return err
		}

		params := sqlcgen.UpdateWatchlistSyncParams{
			ID: id,
			SyncFrequency: pgtype.Interval{
				Microseconds: watchlist.SyncFrequency.Microseconds(),
				Valid:        true,
			},
		}

		if _, err := queries.UpdateWatchlistSync(ctx, params); err != nil {
			if err := handleWriteError(err); err != nil {
				return err
			}

			return handleReadError(err)
		}

		return nil
	})
}
