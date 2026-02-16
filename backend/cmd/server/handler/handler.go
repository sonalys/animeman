package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/cmd/server/middlewares"
	"github.com/sonalys/animeman/cmd/server/ogen"
	"github.com/sonalys/animeman/cmd/server/security"
	"github.com/sonalys/animeman/internal/app/jwt"
	"github.com/sonalys/animeman/internal/app/usecases"
	"github.com/sonalys/animeman/internal/utils/otel"
)

type Handler struct {
	JWTClient *jwt.Client
	Usecases  usecases.Usecases
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
