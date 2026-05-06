package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrAddressNotFound   = errutil.New("ADDRESS_NOT_FOUND", "address not found", codes.NotFound)
	ErrNoChangesProvided = errutil.New("NO_CHANGES", "no changes provided", codes.InvalidArgument)
)
