package handler

import (
	"net/http"

	"github.com/sonalys/animeman/cmd/server/middlewares"
	"github.com/sonalys/animeman/cmd/server/ogen"
	"github.com/sonalys/animeman/cmd/server/security"
	"github.com/sonalys/animeman/internal/app/jwt"
	"github.com/sonalys/animeman/internal/app/usecases"
	"github.com/sonalys/animeman/internal/utils/otel"
)

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
			middlewares.Recoverer,
			middlewares.Logger,
		),
	)
}
