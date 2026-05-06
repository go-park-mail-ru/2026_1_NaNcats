package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrUserNotFound       = errutil.New("USER_NOT_FOUND", "user not found", codes.NotFound)
	ErrEmailAlreadyExists = errutil.New("USER_ALREADY_EXISTS", "user with this email already exists", codes.AlreadyExists)
	ErrInvalidInput       = errutil.New("IVALID_INPUT", "invalid input of data", codes.InvalidArgument)
	ErrInvalidImageExt    = errutil.New("INVALID_EXTENSION", "invalid image extension", codes.InvalidArgument)
	ErrImageNotFound      = errutil.New("IMAGE_NOT_FOUND", "image not found", codes.NotFound)
	ErrNoChangesProvided  = errutil.New("NO_CHANGES", "no changes provided", codes.InvalidArgument)
)
