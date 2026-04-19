package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrInvalidQuantity     = errutil.New("invalid dish quantity", codes.InvalidArgument)
	ErrDishNotFound        = errutil.New("dish not found", codes.NotFound)
	ErrMultipleRestaurants = errutil.New("restaurant is different in cart and dish", codes.InvalidArgument)
)
