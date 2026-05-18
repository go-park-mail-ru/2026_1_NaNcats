package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -destination=mocks/wordle_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase WordleUseCase
type WordleUseCase interface {
	GetDailyState(ctx context.Context, userID int64) (domain.DailyGameState, error)
	MakeGuess(ctx context.Context, userID int64, guess string, idempotencyKey string) (domain.MakeWordleGuessResult, error)
}

type MakeWordleGuessResult struct {
	Status       domain.GameStatus
	GuessResult  domain.WordleGuessResult
	BonusAwarded int64
}

type wordleUseCase struct {
	repo   repository.WordleRepository
	logger logger.Logger
}

func NewWordleUseCase(repo repository.WordleRepository, l logger.Logger) WordleUseCase {
	return &wordleUseCase{
		repo:   repo,
		logger: l,
	}
}

func evaluateGuess(guess, target string) []domain.LetterState {
	states := make([]domain.LetterState, domain.WordleWordLength)
	guessRunes := []rune(guess)
	targetRunes := []rune(target)

	letterCounts := make(map[rune]int)
	for _, r := range targetRunes {
		letterCounts[r]++
	}

	for i := 0; i < domain.WordleWordLength; i++ {
		if guessRunes[i] == targetRunes[i] {
			states[i] = domain.LetterStateCorrect
			letterCounts[guessRunes[i]]--
		}
	}

	for i := 0; i < domain.WordleWordLength; i++ {
		if states[i] == domain.LetterStateCorrect {
			continue
		}

		if count, ok := letterCounts[guessRunes[i]]; ok && count > 0 {
			states[i] = domain.LetterStatePresent
			letterCounts[guessRunes[i]]--
		} else {
			states[i] = domain.LetterStateAbsent
		}
	}

	return states
}

func (u *wordleUseCase) GetDailyState(ctx context.Context, userID int64) (domain.DailyGameState, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", userID))

	now := time.Now()

	targetWord, err := u.repo.GetDailyWord(ctx, now)
	if err != nil {
		return domain.DailyGameState{}, errutil.Internal("failed to get daily word", err)
	}

	game, guessesHistory, err := u.repo.GetGameState(ctx, userID, now)
	if err != nil {
		return domain.DailyGameState{}, errutil.Internal("failed to get game state", err)
	}

	state := domain.DailyGameState{
		Status: domain.GameStatusPlaying,
	}

	if game.Solved {
		state.Status = domain.GameStatusWon
	} else if game.Attempt >= domain.WordleMaxAttempts {
		state.Status = domain.GameStatusLost
	}

	for _, g := range guessesHistory {
		evaluated := evaluateGuess(g.Word, targetWord)
		state.Guesses = append(state.Guesses, domain.WordleGuessResult{
			Word:    g.Word,
			Letters: evaluated,
		})
	}

	return state, nil
}

func (u *wordleUseCase) MakeGuess(ctx context.Context, userID int64, guess string, idempotencyKey string) (domain.MakeWordleGuessResult, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.String("idempotency_key", idempotencyKey),
		attribute.String("guess", guess),
	)

	guess = strings.ToLower(strings.TrimSpace(guess))

	if len([]rune(guess)) != domain.WordleWordLength {
		return domain.MakeWordleGuessResult{}, domain.ErrInvalidWordLength
	}

	now := time.Now()

	targetWord, err := u.repo.GetDailyWord(ctx, now)
	if err != nil {
		return domain.MakeWordleGuessResult{}, errutil.Internal("failed to get daily word", err)
	}

	game, _, err := u.repo.GetGameState(ctx, userID, now)
	if err != nil {
		return domain.MakeWordleGuessResult{}, errutil.Internal("failed to get game state", err)
	}

	if game.FinishedAt != nil {
		return domain.MakeWordleGuessResult{}, domain.ErrGameAlreadyFinished
	}

	exists, err := u.repo.CheckWordExists(ctx, guess)
	if err != nil {
		return domain.MakeWordleGuessResult{}, errutil.Internal("failed to check word existence", err)
	}
	if !exists {
		return domain.MakeWordleGuessResult{}, domain.ErrWordNotInDictionary
	}

	evaluatedStates := evaluateGuess(guess, targetWord)

	isWin := guess == targetWord
	game.Attempt++
	isLoss := !isWin && game.Attempt >= domain.WordleMaxAttempts

	var bonusAwarded int64 = 0
	if isWin {
		bonusAwarded = domain.WordleWinBonus
	}

	guessRecord := domain.WordleGuess{
		UserID:         userID,
		GuessDate:      now,
		AttemptNum:     game.Attempt,
		Word:           guess,
		IdempotencyKey: idempotencyKey,
	}

	err = u.repo.SaveGuessWithTransaction(ctx, guessRecord, isWin, isLoss, bonusAwarded)
	if err != nil {
		if errors.Is(err, domain.ErrIdempotencyConflict) {
			return domain.MakeWordleGuessResult{}, err
		}
		return domain.MakeWordleGuessResult{}, errutil.Internal("failed to save guess", err)
	}

	status := domain.GameStatusPlaying
	if isWin {
		status = domain.GameStatusWon
	} else if isLoss {
		status = domain.GameStatusLost
	}

	return domain.MakeWordleGuessResult{
		Status: status,
		GuessResult: domain.WordleGuessResult{
			Word:    guess,
			Letters: evaluatedStates,
		},
		BonusAwarded: bonusAwarded,
	}, nil
}
