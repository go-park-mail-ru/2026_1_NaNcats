package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrDishNotFound             = errutil.New("DISH_NOT_FOUND", "dish not found", codes.NotFound)
	ErrRestaurantNotFound       = errutil.New("RESTAURANT_NOT_FOUND", "restaurant not found", codes.NotFound)
	ErrInvalidRestaurantBrandID = errutil.New("INVALID_RESTAURANT_BRAND_ID", "invalid restaurant_brand_id", codes.InvalidArgument)
	ErrInvalidInput             = errutil.New("INVALID_INPUT", "invalid input data", codes.InvalidArgument)
	ErrInvalidImageExt          = errutil.New("INVALID_IMAGE_EXT", "invalid image extension", codes.InvalidArgument)
	ErrUnauthorized             = errutil.New("UNAUTHORIZED", "unauthorized access", codes.Unauthenticated)
	ErrPermissionDenied         = errutil.New("PERMISSION_DENIED", "permission denied: you do not own this restaurant", codes.PermissionDenied)
)
