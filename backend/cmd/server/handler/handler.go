package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/cmd/server/middlewares"
	"github.com/sonalys/animeman/cmd/server/ogen"
	"github.com/sonalys/animeman/cmd/server/security"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/app/jwt"
	"github.com/sonalys/animeman/internal/app/usecases"
	"github.com/sonalys/animeman/internal/domain/authentication"
	"github.com/sonalys/animeman/internal/domain/indexing"
	"github.com/sonalys/animeman/internal/domain/transfer"
	"github.com/sonalys/animeman/internal/utils/otel"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
	"google.golang.org/grpc/codes"
)

type Handler struct {
	JWTClient *jwt.Client
	Usecases  usecases.Usecases
}

func (h *Handler) SetupGet(ctx context.Context) (*ogen.SetupGetOK, error) {
	return &ogen.SetupGetOK{
		CompletedSteps: []ogen.SetupSteps{},
		MissingSteps:   []ogen.SetupSteps{},
	}, nil
}

func (h *Handler) TestIndexingClientConfiguration(ctx context.Context, req *ogen.IndexerConfig) error {
	userID, err := security.GetIdentity(ctx)
	if err != nil {
		return err
	}

	b := indexing.NewClientBuilder().
		WithType(func() indexing.ClientType {
			switch req.Type {
			case ogen.IndexerClientTypeProwlarr:
				return indexing.IndexerTypeProwlarr
			default:
				return indexing.IndexerTypeUnknown
			}
		}()).
		WithAddress(req.Hostname).
		WithOwner(userID).
		WithAuth(func() authentication.Authentication {
			switch req.Auth.Type {
			case ogen.AuthenticationTypeApiKey:
				auth := req.Auth.OneOf.AuthenticationAPIKey
				return authentication.NewAPIKeyAuthentication(auth.Key)
			default:
				return authentication.Authentication{}
			}
		}())

	if err := h.Usecases.TestIndexingClientBuilder(ctx, b); err != nil {
		if errors.Is(err, authentication.ErrUnsupportedAuthentication) {
			return apperr.NewFieldError(apperr.FieldErrorCodeUnsupported, "auth.type")
		}

		if apperr.Code(err) == codes.Unauthenticated {
			switch req.Auth.Type {
			case ogen.AuthenticationTypeUserPassword:
				validation := apperr.NewFormValidation(
					apperr.NewFieldError(apperr.FieldErrorCodeInvalid, "auth.username"),
					apperr.NewFieldError(apperr.FieldErrorCodeInvalid, "auth.password"),
				)

				return apperr.NewPublicError(validation.Validate(), "username/password mismatch")
			case ogen.AuthenticationTypeApiKey:
				return apperr.NewFieldError(apperr.FieldErrorCodeInvalid, "auth.key")
			}
		}

		if apperr.Code(err) == codes.InvalidArgument {
			return apperr.NewFieldError(apperr.FieldErrorCodeInvalid, "hostname")
		}

		return err
	}

	return nil
}

func (h *Handler) TestTransferClientConfiguration(ctx context.Context, req *ogen.TransferClientConfig) error {
	userID, err := security.GetIdentity(ctx)
	if err != nil {
		return err
	}

	b := transfer.NewClientBuilder().
		WithType(func() transfer.ClientType {
			switch req.Type {
			case ogen.TransferClientTypeQbittorrent:
				return transfer.ClientTypeQBittorrent
			default:
				return transfer.ClientTypeUnknown
			}
		}()).
		WithAddress(req.Hostname).
		WithOwner(userID).
		WithAuth(func() authentication.Authentication {
			switch req.Auth.Type {
			case ogen.AuthenticationTypeUserPassword:
				auth := req.Auth.OneOf.AuthenticationUserPassword
				return authentication.NewUserPasswordAuthentication(auth.Username, []byte(auth.Password))
			default:
				return authentication.Authentication{}
			}
		}())

	if err := h.Usecases.TestTransferClientBuilder(ctx, b); err != nil {
		if errors.Is(err, authentication.ErrUnsupportedAuthentication) {
			return apperr.NewFieldError(apperr.FieldErrorCodeUnsupported, "auth.type")
		}

		if apperr.Code(err) == codes.Unauthenticated {
			switch req.Auth.Type {
			case ogen.AuthenticationTypeUserPassword:
				validation := apperr.NewFormValidation(
					apperr.NewFieldError(apperr.FieldErrorCodeInvalid, "auth.username"),
					apperr.NewFieldError(apperr.FieldErrorCodeInvalid, "auth.password"),
				)

				return apperr.NewPublicError(validation.Validate(), "username/password mismatch")
			case ogen.AuthenticationTypeApiKey:
				return apperr.NewFieldError(apperr.FieldErrorCodeInvalid, "auth.key")
			}
		}

		if apperr.Code(err) == codes.InvalidArgument {
			return apperr.NewFieldError(apperr.FieldErrorCodeInvalid, "hostname")
		}

		return err
	}

	return nil
}

func (h *Handler) IndexingClientsGet(ctx context.Context) ([]ogen.Indexer, error) {
	userID, err := security.GetIdentity(ctx)
	if err != nil {
		return nil, err
	}

	response, err := h.Usecases.ListIndexers(ctx, userID)
	if err != nil {
		return nil, err
	}

	return sliceutils.Map(response, func(from indexing.Client) ogen.Indexer {
		return ogen.Indexer{
			ID:       uuid.UUID(from.ID.UUID),
			Type:     ogen.IndexerClientType(from.Type.String()),
			Hostname: from.Address,
		}
	}), nil
}

func (h *Handler) IndexingClientsPost(ctx context.Context, req *ogen.IndexerConfig) (*ogen.IndexingClientsPostCreated, error) {
	userID, err := security.GetIdentity(ctx)
	if err != nil {
		return nil, err
	}

	client, err := h.Usecases.CreateIndexer(ctx, usecases.CreateIndexerArgs{
		UserID: userID,
		URL:    req.Hostname,
		Type: func() indexing.ClientType {
			switch req.Type {
			case ogen.IndexerClientTypeProwlarr:
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

	return &ogen.IndexingClientsPostCreated{
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
		ogen.WithNotFound(notFound),
		ogen.WithMiddleware(
			middlewares.Logger,
			middlewares.Recoverer,
		),
	)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	response := ogen.ErrorResponse{
		Details:     ogen.NewOptString("not found"),
		FieldErrors: nil,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Ctx(r.Context()).Err(err).Msg("failed to encode error response")
	}
}
