package domain

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

var (
	ErrSessionNotFound = errutil.New("SESSION_NOT_FOUND", "session not found", codes.NotFound)
	ErrSessionExpired  = errutil.New("SESSION_NOT_EXPIRED", "session expired", codes.Unauthenticated)
)
