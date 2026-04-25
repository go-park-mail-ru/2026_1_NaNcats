package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrTicketNotFound      = errutil.New("TICKET_NOT_FOUND", "support ticket not found", codes.NotFound)
	ErrCategoryNotFound    = errutil.New("CATEGORY_NOT_FOUND", "support category not found", codes.NotFound)
	ErrInvalidRating       = errutil.New("INVALID_RATING", "rating must be between 1 and 5", codes.InvalidArgument)
	ErrUnauthorized        = errutil.New("UNAUTHORIZED", "guest_id or client_id is required", codes.Unauthenticated)
	ErrInvalidStatusInput  = errutil.New("INVALID_STATUS_INPUT", "status must be online or offline", codes.InvalidArgument)
	ErrInvalidMessageInput = errutil.New("INVALID_MSG_INPUT", "status must be online or offline", codes.InvalidArgument)
	ErrPermissionDenied    = errutil.New("PERMISSION_DENIED", "access denied", codes.PermissionDenied)
	ErrInvalidState        = errutil.New("INVALID_STATE", "only closed tickets can be rated", codes.FailedPrecondition)
	ErrInvalidRatingInput  = errutil.New("INVALID_INPUT", "rating must be between 1 and 5", codes.InvalidArgument)
)
