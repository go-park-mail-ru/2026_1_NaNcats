package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrDishNotFound       = errutil.New("dish not found", codes.NotFound)
	ErrRestaurantNotFound = errutil.New("restaurant not found", codes.NotFound)
)
