package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/cmd/server/middlewares"
	"github.com/sonalys/animeman/cmd/server/ogen"
	"github.com/sonalys/animeman/cmd/server/security"
	"github.com/sonalys/animeman/internal/app/jwt"
	"github.com/sonalys/animeman/internal/app/usecases"
	"github.com/sonalys/animeman/internal/domain/authentication"
	"github.com/sonalys/animeman/internal/domain/indexing"
	"github.com/sonalys/animeman/internal/utils/otel"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
)

type Handler struct {
	JWTClient *jwt.Client
	Usecases  usecases.Usecases
}

// IndexersGet implements [ogen.Handler].
func (h *Handler) IndexersGet(ctx context.Context) ([]ogen.Indexer, error) {
	userID, err := security.GetIdentity(ctx)
	if err != nil {
		return nil, err
	}

	response, err := h.Usecases.ListIndexers(ctx, userID)
	if err != nil {
		return nil, err
	}

	return sliceutils.Map(response, func(from indexing.IndexerClient) ogen.Indexer {
		return ogen.Indexer{
			ID:   uuid.UUID(from.ID.UUID),
			Type: ogen.IndexerType(from.Type.String()),
			URL:  from.Address,
		}
	}), nil
}

// IndexersPost implements [ogen.Handler].
func (h *Handler) IndexersPost(ctx context.Context, req *ogen.IndexerConfig) (*ogen.IndexersPostCreated, error) {
	userID, err := security.GetIdentity(ctx)
	if err != nil {
		return nil, err
	}

	client, err := h.Usecases.CreateIndexer(ctx, usecases.CreateIndexerArgs{
		UserID: userID,
		URL:    req.URL,
		Type: func() indexing.IndexerType {
			switch req.Type {
			case ogen.IndexerTypeProwlarr:
				return indexing.IndexerTypeProwlarr
			default:
				return indexing.IndexerTypeUnknown
			}
		}(),
		Auth: func() authentication.Authentication {
			switch req.Auth.Type {
			case ogen.AuthenticationTypeApiKey:
				auth := req.Auth.OneOf.AuthenticationAPIKey

				return authentication.NewAPIKeyAuthentication(auth.Key)
			case ogen.AuthenticationTypeUserPassword:
				auth := req.Auth.OneOf.AuthenticationUserPassword

				return authentication.NewUserPasswordAuthentication(auth.Username, []byte(auth.Password))
			default:
				return authentication.Authentication{Type: authentication.AuthenticationTypeUnknown}
			}
		}(),
	})

	return &ogen.IndexersPostCreated{
		ID: uuid.UUID(client.ID.UUID),
	}, nil
}

func New(
	jwtClient *jwt.Client,
	usecases usecases.Usecases,
) (http.Handler, error) {
	securityHandler := security.NewHandler(jwtClient)

	handler := &Handler{
		Usecases:  usecases,
		JWTClient: jwtClient,
	}

	return ogen.NewServer(handler, securityHandler,
		ogen.WithPathPrefix("/api/v1"),
		ogen.WithTracerProvider(otel.Provider.TracerProvider()),
		ogen.WithErrorHandler(validationErrorHandler(handler)),
		ogen.WithNotFound(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)

			response := ogen.ErrorResponse{
				Details:     ogen.NewOptString("not found"),
				FieldErrors: nil,
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				log.Error().Ctx(r.Context()).Err(err).Msg("failed to encode error response")
			}
		}),
		ogen.WithMiddleware(
			middlewares.Logger,
			middlewares.Recoverer,
		),
	)
}
