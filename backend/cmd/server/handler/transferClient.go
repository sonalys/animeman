package handler

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sonalys/animeman/cmd/server/ogen"
	"github.com/sonalys/animeman/cmd/server/security"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/app/usecases"
	"github.com/sonalys/animeman/internal/domain/authentication"
	"github.com/sonalys/animeman/internal/domain/transfer"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
	"google.golang.org/grpc/codes"
)

func (h *Handler) TransferClientsGet(ctx context.Context) ([]ogen.TransferClient, error) {
	userID, err := security.GetIdentity(ctx)
	if err != nil {
		return nil, err
	}

	response, err := h.Usecases.ListTransferClients(ctx, userID)
	if err != nil {
		return nil, err
	}

	return sliceutils.Map(response, func(from transfer.Client) ogen.TransferClient {
		return ogen.TransferClient{
			ID:       uuid.UUID(from.ID.UUID),
			Type:     ogen.TransferClientType(from.Type.String()),
			Hostname: from.Address,
		}
	}), nil
}

func (h *Handler) TransferClientsPost(ctx context.Context, req *ogen.TransferClientConfig) (*ogen.TransferClientsPostCreated, error) {
	userID, err := security.GetIdentity(ctx)
	if err != nil {
		return nil, err
	}

	client, err := h.Usecases.CreateTransferClient(ctx, usecases.CreateTransferClientArgs{
		UserID: userID,
		URL:    req.Hostname,
		Type: func() transfer.ClientType {
			switch req.Type {
			case ogen.TransferClientTypeQbittorrent:
				return transfer.ClientTypeQBittorrent
			default:
				return transfer.ClientTypeUnknown
			}
		}(),
		Auth: func() authentication.Authentication {
			switch req.Auth.OneOf.Type {
			case ogen.AuthenticationAPIKeyAuthenticationSum:
				auth := req.Auth.OneOf.AuthenticationAPIKey

				return authentication.NewAPIKeyAuthentication(auth.Key)
			case ogen.AuthenticationUserPasswordAuthenticationSum:
				auth := req.Auth.OneOf.AuthenticationUserPassword

				return authentication.NewUserPasswordAuthentication(auth.Username, []byte(auth.Password))
			default:
				return authentication.Authentication{Type: authentication.AuthenticationTypeUnknown}
			}
		}(),
	})

	return &ogen.TransferClientsPostCreated{
		ID: uuid.UUID(client.ID.UUID),
	}, nil
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
			switch req.Auth.OneOf.Type {
			case ogen.AuthenticationUserPasswordAuthenticationSum:
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
			switch req.Auth.OneOf.Type {
			case ogen.AuthenticationUserPasswordAuthenticationSum:
				validation := apperr.NewFormValidation(
					apperr.NewFieldError(apperr.FieldErrorCodeInvalid, "auth.username"),
					apperr.NewFieldError(apperr.FieldErrorCodeInvalid, "auth.password"),
				)

				return apperr.NewPublicError(validation.Validate(), "username/password mismatch")
			case ogen.AuthenticationAPIKeyAuthenticationSum:
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
