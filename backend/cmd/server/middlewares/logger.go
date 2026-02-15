package middlewares

import (
	"github.com/ogen-go/ogen/middleware"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/app/apperr"
)

func Logger(req middleware.Request, next middleware.Next) (resp middleware.Response, err error) {
	ctx := req.Context

	logger := log.Info().
		Ctx(ctx).
		Str("operationId", req.OperationID)

	logger.Msg("request received")

	resp, err = next(req)
	if err == nil {
		if code := apperr.Code(err); code != 0 {
			logger = logger.Stringer("errorCode", code)
		}
	}

	logger.Msg("request responded")

	return
}
