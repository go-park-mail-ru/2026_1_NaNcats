package repository

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
)

//go:generate mockgen -destination=mocks/wordle_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository WordleRepository
type WordleRepository interface {
	GetDailyWord(ctx context.Context, date time.Time) (string, error)
	CheckWordExists(ctx context.Context, word string) (bool, error)
	GetGameState(ctx context.Context, userID int64, date time.Time) (domain.WordleGame, []domain.WordleGuess, error)
	SaveGuessWithTransaction(ctx context.Context, guess domain.WordleGuess, isWin, isLoss bool) error
	GetCurrentStreak(ctx context.Context, userID int64) (int32, error)
	CountWins(ctx context.Context, userID int64) (int32, error)
}
