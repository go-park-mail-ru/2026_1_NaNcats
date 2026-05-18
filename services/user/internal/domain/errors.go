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

	ErrGameAlreadyFinished = errutil.New("GAME_ALREADY_FINISHED", "the game is already finished for today", codes.FailedPrecondition)
	ErrWordNotInDictionary = errutil.New("WORD_NOT_IN_DICTIONARY", "word is not in the dictionary", codes.InvalidArgument)
	ErrInvalidWordLength   = errutil.New("INVALID_WORD_LENGTH", "word length must be exactly 5 letters", codes.InvalidArgument)
	ErrMaxAttemptsReached  = errutil.New("MAX_ATTEMPTS_REACHED", "max attempts reached", codes.FailedPrecondition)
	ErrDailyWordNotFound   = errutil.New("DAILY_WORD_NOT_FOUND", "daily word not set for today", codes.Internal)
	ErrIdempotencyConflict = errutil.New("IDEMPOTENCY_CONFLICT", "request with this idempotency key already processed", codes.AlreadyExists)
)
