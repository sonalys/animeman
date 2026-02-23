package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/cmd/server/ogen"
	"github.com/sonalys/animeman/cmd/server/security"
	"github.com/sonalys/animeman/internal/app/jwt"
	"github.com/sonalys/animeman/internal/app/usecases"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	JWTClient *jwt.Client
	Usecases  usecases.Usecases
}

func (h *Handler) SetupGet(ctx context.Context) (*ogen.SetupGetOK, error) {
	userID, err := security.GetIdentity(ctx)
	if err != nil {
		return nil, err
	}

	status, err := h.Usecases.GetOnboardingStatus(ctx, userID)
	if err != nil {
		return nil, err
	}

	remapStatus := func(from usecases.SetupStep) ogen.SetupSteps {
		switch from {
		case usecases.SetupStepIndexingClient:
			return ogen.SetupStepsIndexing
		case usecases.SetupStepTransferClient:
			return ogen.SetupStepsTransfer
		case usecases.SetupStepWatchlistSetupStep:
			return ogen.SetupStepsWatchlist
		default:
			return ""
		}
	}

	return &ogen.SetupGetOK{
		IsCompleted:    status.IsSetupCompleted,
		CompletedSteps: sliceutils.Map(status.CompletedSteps, remapStatus),
		MissingSteps:   sliceutils.Map(status.MissingSteps, remapStatus),
	}, nil
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	span := trace.SpanFromContext(r.Context())

	var respTraceID uuid.UUID

	if traceID := span.SpanContext().TraceID(); traceID.IsValid() {
		respTraceID = uuid.UUID(traceID)
	}

	response := ogen.ErrorResponse{
		TraceID:     ogen.NewOptUUID(respTraceID),
		Details:     ogen.NewOptString("not found"),
		FieldErrors: nil,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Ctx(r.Context()).Err(err).Msg("failed to encode error response")
	}
}
