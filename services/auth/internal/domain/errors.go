package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrSessionNotFound = errutil.New("session not found", codes.NotFound)
	ErrSessionExpired  = errutil.New("session expired", codes.Unauthenticated)
)
