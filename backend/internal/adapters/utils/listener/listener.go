package listener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/ports"
)

func Listen(ctx context.Context, conn *pgxpool.Pool, topic string) iter.Seq[ports.RepositoryNotification] {
	openNotifier := func() (*pgxpool.Conn, error) {
		conn, err := conn.Acquire(ctx)
		if err != nil {
			return nil, err
		}

		log.Debug().
			Str("topic", topic).
			Msg("Opening listener on topic")

		if _, err = conn.Exec(ctx, "LISTEN "+topic); err != nil {
			return nil, fmt.Errorf("listening to topic: %w", err)
		}

		return conn, nil
	}

	return func(yield func(ports.RepositoryNotification) bool) {
		conn, err := openNotifier()
		if err != nil {
			return
		}

		for {
			const reconnectDelay = 5 * time.Second

			if conn.Conn().IsClosed() {
				log.Debug().
					Str("topic", topic).
					Dur("delay", reconnectDelay).
					Msgf("Listener was closed, reconnecting")

				conn, err = openNotifier()
				if err != nil {
					select {
					case <-ctx.Done():
						return
					case <-time.After(reconnectDelay):
						continue
					}
				}
			}

			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				continue
			}

			type CollectionNotification struct {
				Action    string    `json:"action"`
				ID        string    `json:"id"`
				Table     string    `json:"table"`
				ChangedAt time.Time `json:"changed_at"`
			}

			var payload CollectionNotification
			if err = json.Unmarshal([]byte(notification.Payload), &payload); err != nil {
				fmt.Printf("Error decoding JSON: %v\n", err)
				continue
			}

			log.Debug().
				Str("topic", topic).
				Any("payload", payload).
				Msg("Received notification")

			if !yield(ports.RepositoryNotification{
				ID:        shared.ParseStringID[shared.ID](payload.ID),
				ChangedAt: payload.ChangedAt,
				Action: func() ports.RepositoryAction {
					switch payload.Action {
					case "INSERT":
						return ports.RepositoryActionCreate
					case "UPDATE":
						return ports.RepositoryActionUpdate
					case "DELETE":
						return ports.RepositoryActionDelete
					default:
						return ports.RepositoryActionUnknown
					}
				}(),
			}) {
				return
			}
		}
	}
}
