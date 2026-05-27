package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrUnauthorized       = errutil.New("UNAUTHORIZED", "unauthorized access: session not found", codes.Unauthenticated)
	ErrPermissionDenied   = errutil.New("PERMISSION_DENIED", "permission denied: you do not own this restaurant", codes.PermissionDenied)
	ErrRestaurantNotFound = errutil.New("RESTAURANT_NOT_FOUND", "restaurant brand not found", codes.NotFound)
)
