package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrDishNotFound       = errutil.New("DISH_NOT_FOUND", "dish not found", codes.NotFound)
	ErrRestaurantNotFound = errutil.New("RESTAURANT_NOT_FOUND", "restaurant not found", codes.NotFound)
)
