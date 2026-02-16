package middlewares

import (
	"github.com/ogen-go/ogen/middleware"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/app/apperr"
)

func Logger(req middleware.Request, next middleware.Next) (resp middleware.Response, err error) {
	logger := log.
		With().
		Str("operationId", req.OperationID).
		Logger()

	logger.Info().Ctx(req.Context).Msg("Request received")

	resp, err = next(req)
	if err == nil {
		if code := apperr.Code(err); code != 0 {
			logger = logger.
				With().
				Stringer("errorCode", code).
				Logger()
		}
	}

	logger.Info().Ctx(req.Context).Msg("Request responded")

	return
}
