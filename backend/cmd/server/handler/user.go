package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sonalys/animeman/cmd/server/ogen"
	"github.com/sonalys/animeman/cmd/server/security"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/app/jwt"
	"github.com/sonalys/animeman/internal/domain/users"
)

func (h *Handler) AuthenticationLogin(ctx context.Context, req *ogen.AuthenticationLoginReq) (*ogen.AuthenticationLoginOK, error) {
	userID, err := h.Usecases.Login(ctx, req.Username, []byte(req.Password))
	if err != nil {
		return nil, err
	}

	stringifiedToken, err := h.JWTClient.Encode(&jwt.Token{
		UserID: *userID,
		Exp:    time.Now().Add(24 * time.Hour).Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &ogen.AuthenticationLoginOK{
		SetCookie: fmt.Sprintf("SESSION_ID=%s; Path=/; HttpOnly; SameSite=Strict", stringifiedToken),
	}, nil
}

func (h *Handler) AuthenticationWhoAmI(ctx context.Context) (*ogen.AuthenticationWhoAmIOK, error) {
	userID, err := security.GetIdentity(ctx)
	if err != nil {
		return nil, err
	}

	return &ogen.AuthenticationWhoAmIOK{
		UserID: uuid.UUID(userID.UUID),
	}, nil
}

func (h *Handler) RegisterUser(ctx context.Context, req *ogen.UserRegistration) (ogen.RegisterUserRes, error) {
	user, err := h.Usecases.RegisterUser(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, users.ErrUniqueUsername) {
			return nil, apperr.NewFieldError(apperr.FieldErrorCodeAlreadyExists, "username", "conflict: %s", err)
		}

		return nil, err
	}

	token := &jwt.Token{
		UserID: user.ID,
		Exp:    time.Now().Add(24 * time.Hour).Unix(),
	}

	stringifiedToken, err := h.JWTClient.Encode(token)
	if err != nil {
		return nil, err
	}

	return &ogen.RegisterUserCreatedHeaders{
		SetCookie: fmt.Sprintf("SESSION_ID=%s; Path=/; HttpOnly; SameSite=Strict", stringifiedToken),
		Response: ogen.RegisterUserCreated{
			ID: uuid.UUID(user.ID.UUID),
		},
	}, nil
}
