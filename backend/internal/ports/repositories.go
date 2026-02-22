package ports

import (
	"context"
	"time"

	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/domain/indexing"
	"github.com/sonalys/animeman/internal/domain/orchestration"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/domain/transfer"
	"github.com/sonalys/animeman/internal/domain/users"
	"github.com/sonalys/animeman/internal/domain/watchlists"
)

type (
	UpdateHandler[T any] = func(*T) error

	ListOptions struct {
		PageSize int32
		Cursor   shared.ID
	}

	UserRepository interface {
		Create(ctx context.Context, user *users.User) error
		Get(ctx context.Context, id shared.UserID) (*users.User, error)
		GetByUsername(ctx context.Context, username string) (*users.User, error)
		Update(ctx context.Context, id shared.UserID, updateHandler UpdateHandler[users.User]) error
		Delete(ctx context.Context, id shared.UserID) error
	}

	IndexerClientRepository interface {
		Create(ctx context.Context, client *indexing.Client) error
		List(ctx context.Context) ([]indexing.Client, error)
		ListByOwner(ctx context.Context, id shared.UserID) ([]indexing.Client, error)
		Update(ctx context.Context, id indexing.IndexerID, updateHandler UpdateHandler[indexing.Client]) error
		Delete(ctx context.Context, id indexing.IndexerID) error
	}

	TransferClientRepository interface {
		Create(ctx context.Context, client *transfer.Client) error
		List(ctx context.Context) ([]transfer.Client, error)
		ListByOwner(ctx context.Context, id shared.UserID) ([]transfer.Client, error)
		Update(ctx context.Context, id transfer.ClientID, updateHandler UpdateHandler[transfer.Client]) error
		Delete(ctx context.Context, id transfer.ClientID) error
	}

	CollectionRepository interface {
		Create(ctx context.Context, collection *collections.Collection) error
		ListByOwner(ctx context.Context, id shared.UserID) ([]collections.Collection, error)
		Update(ctx context.Context, id collections.CollectionID, updateHandler UpdateHandler[collections.Collection]) error
		Delete(ctx context.Context, id collections.CollectionID) error
	}

	QualityProfileRepository interface {
		Create(ctx context.Context, qualityProfile *collections.QualityProfile) error
		List(ctx context.Context) ([]collections.QualityProfile, error)
		Update(ctx context.Context, id collections.QualityProfileID, updateHandler UpdateHandler[collections.QualityProfile]) error
		Delete(ctx context.Context, id collections.QualityProfileID) error
	}

	MediaRepository interface {
		Create(ctx context.Context, media *collections.Media) error
		ListByCollection(ctx context.Context, id collections.CollectionID, opts ListOptions) ([]collections.Media, error)
		Update(ctx context.Context, id collections.MediaID, updateHandler UpdateHandler[collections.Media]) error
		Delete(ctx context.Context, id collections.MediaID) error
	}

	SeasonRepository interface {
		Create(ctx context.Context, season *collections.Season) error
		ListByMedia(ctx context.Context, id collections.MediaID) ([]collections.Season, error)
		Update(ctx context.Context, id collections.SeasonID, updateHandler UpdateHandler[collections.Season]) error
		Delete(ctx context.Context, id collections.SeasonID) error
	}

	EpisodeRepository interface {
		Create(ctx context.Context, episode *collections.Episode) error
		ListBySeason(ctx context.Context, id collections.SeasonID) ([]collections.Episode, error)
		Update(ctx context.Context, id collections.EpisodeID, updateHandler UpdateHandler[collections.Episode]) error
		Delete(ctx context.Context, id collections.EpisodeID) error
	}

	WatchlistRepository interface {
		Create(ctx context.Context, watchlist *watchlists.Watchlist) error
		CreateEntry(ctx context.Context, entry *watchlists.WatchlistEntry) error
		List(ctx context.Context) ([]watchlists.Watchlist, error)
		ListByOwner(ctx context.Context, id shared.UserID) ([]watchlists.Watchlist, error)
		ListEntries(ctx context.Context, id watchlists.WatchlistID) ([]watchlists.WatchlistEntry, error)
		Update(ctx context.Context, id watchlists.WatchlistID, updateHandler UpdateHandler[watchlists.Watchlist]) error
		Delete(ctx context.Context, id watchlists.WatchlistID) error
		DeleteEntry(ctx context.Context, id watchlists.WatchlistEntryID) error
	}

	TaskRepository interface {
		// Task Lifecycle
		CreateTask(ctx context.Context, t *orchestration.Task) error
		ClaimTask(ctx context.Context, timeout time.Duration) (*orchestration.Task, error)
		MarkCompleted(ctx context.Context, id orchestration.TaskID) error
		MarkFailed(ctx context.Context, id orchestration.TaskID, nextRetry time.Time) error

		// Logging & Telemetry
		AddLog(ctx context.Context, entry *orchestration.TaskLog) error

		// Querying & Pagination
		ListTasks(ctx context.Context, lastID orchestration.TaskID, pageSize int32) ([]orchestration.Task, error)
		ListTaskLogs(ctx context.Context, taskID orchestration.TaskID, lastID orchestration.TaskLogID, pageSize int32) ([]orchestration.TaskLog, error)

		// Maintenance
		RotateLogs(ctx context.Context, retention time.Duration, maxLogs int32) error
	}
)
