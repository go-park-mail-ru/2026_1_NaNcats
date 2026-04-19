package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrUserNotFound       = errutil.New("user not found", codes.NotFound)
	ErrEmailAlreadyExists = errutil.New("user with this email already exists", codes.AlreadyExists)
	ErrInvalidInput       = errutil.New("invalid input of data", codes.InvalidArgument)
	ErrInvalidImageExt    = errutil.New("invalid image extension", codes.InvalidArgument)
	ErrImageNotFound      = errutil.New("image not found", codes.NotFound)
	ErrNoChangesProvided  = errutil.New("no changes provided", codes.InvalidArgument)
)
