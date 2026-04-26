package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrInvalidQuantity     = errutil.New("INVALID_QUANTITY", "invalid dish quantity", codes.InvalidArgument)
	ErrDishNotFound        = errutil.New("DISH_NOT_FOUND", "dish not found", codes.NotFound)
	ErrMultipleRestaurants = errutil.New("MULTIPLE_RESTAURANTS", "restaurant is different in cart and dish", codes.InvalidArgument)

	ErrForbidden       = errutil.New("FORBIDDEN", "user has no rights for this operation", codes.PermissionDenied)
	ErrUnassignedItems = errutil.New("UNASSIGNED_ITEMS", "cart contains orphaned items", codes.FailedPrecondition)
	ErrInviteExpired   = errutil.New("INVITE_EXPIRED", "invite token has expired", codes.NotFound)
	ErrUserNotInCart   = errutil.New("USER_NOT_IN_CART", "user is not a member of this cart", codes.FailedPrecondition)
)
