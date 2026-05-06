package response

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Переводит grpc ошибку в ошибку http + json
func WriteError(w http.ResponseWriter, grpcErr error) {
	st, ok := status.FromError(grpcErr)
	if !ok {
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var httpStatus int
	switch st.Code() {
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.DeadlineExceeded:
		httpStatus = http.StatusGatewayTimeout
	default:
		httpStatus = http.StatusInternalServerError
	}

	Error(w, httpStatus, st.Message())
}
