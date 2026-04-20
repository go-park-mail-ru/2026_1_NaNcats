package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrInvalidQuantity     = errutil.New("INVALID_QUANTITY", "invalid dish quantity", codes.InvalidArgument)
	ErrDishNotFound        = errutil.New("DISH_NOT_FOUND", "dish not found", codes.NotFound)
	ErrMultipleRestaurants = errutil.New("MULTIPLE_RESTAURANTS", "restaurant is different in cart and dish", codes.InvalidArgument)
)
