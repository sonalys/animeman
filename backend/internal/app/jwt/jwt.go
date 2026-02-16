package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain/shared"
	"google.golang.org/grpc/codes"
)

type (
	Client struct {
		jwtSignKey []byte
	}

	Token struct {
		UserID shared.UserID
		Exp    int64
	}
)

const (
	ErrInvalidToken apperr.StringError = "token is invalid"
)

const (
	userClaim = "user"
	expClaim  = "exp"
)

func NewClient(jwtSignKey []byte) *Client {
	return &Client{
		jwtSignKey: jwtSignKey,
	}
}

func (c *Client) Decode(tokenString string) (*Token, error) {
	var claims jwt.MapClaims

	keyFunc := func(token *jwt.Token) (any, error) {
		return c.jwtSignKey, nil
	}

	supportedMethods := []string{
		jwt.SigningMethodHS256.Alg(),
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		keyFunc,
		jwt.WithValidMethods(supportedMethods),
	)
	if err != nil || !token.Valid {
		return nil, apperr.New(err, codes.Unauthenticated, "token could not be parsed")
	}

	userID, ok := claims[userClaim].(string)
	if !ok {
		return nil, apperr.New(ErrInvalidToken, codes.Unauthenticated, "userID is missing")
	}

	userUUID := shared.ParseStringID[shared.UserID](userID)

	exp, ok := claims[expClaim].(float64)
	if !ok {
		return nil, apperr.New(ErrInvalidToken, codes.Unauthenticated, "exp is missing")
	}

	identity := &Token{
		UserID: userUUID,
		Exp:    int64(exp),
	}

	return identity, nil
}

func (c *Client) Encode(identity *Token) (string, error) {
	claims := jwt.MapClaims{
		userClaim: identity.UserID.String(),
		expClaim:  identity.Exp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	stringifiedToken, err := token.SignedString(c.jwtSignKey)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return stringifiedToken, nil
}
