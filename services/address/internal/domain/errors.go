package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrAddressNotFound   = errutil.New("address not found", codes.NotFound)
	ErrNoChangesProvided = errutil.New("no changes provided", codes.InvalidArgument)
)
