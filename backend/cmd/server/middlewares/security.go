package middlewares

import (
	"context"
	"errors"

	"github.com/sonalys/animeman/cmd/server/ogen"
	"github.com/sonalys/animeman/internal/domain/shared"
)

type (
	contextKey string

	TokenDecoder interface {
		Decode(jwt string) (*shared.UserID, error)
	}

	SecurityHandler struct {
		controller TokenDecoder
	}
)

const identityContextKey = contextKey("identity-key")

func NewSecurityHandler(c TokenDecoder) *SecurityHandler {
	return &SecurityHandler{
		controller: c,
	}
}

func GetIdentity(ctx context.Context) (*shared.UserID, error) {
	identity, ok := ctx.Value(identityContextKey).(*shared.UserID)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	return identity, nil
}

func (h *SecurityHandler) HandleCookieAuth(ctx context.Context, operationName ogen.OperationName, auth ogen.CookieAuth) (context.Context, error) {
	identity, err := h.controller.Decode(auth.APIKey)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, identityContextKey, identity), nil
}
