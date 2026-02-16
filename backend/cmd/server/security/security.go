package security

import (
	"context"

	"github.com/sonalys/animeman/cmd/server/ogen"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/app/jwt"
	"github.com/sonalys/animeman/internal/domain/shared"
	"google.golang.org/grpc/codes"
)

type (
	ctxKey struct{}

	SecurityHandler struct {
		client *jwt.Client
	}
)

var errUnauthenticated apperr.Error = apperr.New(nil, codes.Unauthenticated, "unauthenticated")

func NewHandler(c *jwt.Client) *SecurityHandler {
	return &SecurityHandler{
		client: c,
	}
}

func GetIdentity(ctx context.Context) (shared.UserID, error) {
	identity, ok := ctx.Value(ctxKey{}).(shared.UserID)
	if !ok {
		return shared.UserID{}, errUnauthenticated
	}

	return identity, nil
}

func (h *SecurityHandler) HandleCookieAuth(ctx context.Context, operationName ogen.OperationName, auth ogen.CookieAuth) (context.Context, error) {
	identity, err := h.client.Decode(auth.APIKey)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, ctxKey{}, identity.UserID), nil
}
