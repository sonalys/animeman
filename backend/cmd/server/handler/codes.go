package handler

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

// GRPCCodeToHTTP maps a gRPC status code to its corresponding HTTP status code
// based on the official gRPC specification.
func GRPCCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout // 408
	case codes.Unknown:
		return http.StatusInternalServerError // 500
	case codes.InvalidArgument:
		return http.StatusBadRequest // 400
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout // 504
	case codes.NotFound:
		return http.StatusNotFound // 404
	case codes.AlreadyExists:
		return http.StatusConflict // 409
	case codes.PermissionDenied:
		return http.StatusForbidden // 403
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests // 429
	case codes.FailedPrecondition:
		return http.StatusBadRequest // 400
	case codes.Aborted:
		return http.StatusConflict // 409
	case codes.OutOfRange:
		return http.StatusBadRequest // 400
	case codes.Unimplemented:
		return http.StatusNotImplemented // 501
	case codes.Internal:
		return http.StatusInternalServerError // 500
	case codes.Unavailable:
		return http.StatusServiceUnavailable // 503
	case codes.DataLoss:
		return http.StatusInternalServerError // 500
	case codes.Unauthenticated:
		return http.StatusUnauthorized // 401
	default:
		return http.StatusInternalServerError // 500
	}
}
