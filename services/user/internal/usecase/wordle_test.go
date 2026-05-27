package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_evaluateGuess(t *testing.T) {
	tests := []struct {
		name   string
		guess  string
		target string
		want   []domain.LetterState
	}{
		{
			name:   "Полное совпадение",
			guess:  "apple",
			target: "apple",
			want: []domain.LetterState{
				domain.LetterStateCorrect,
				domain.LetterStateCorrect,
				domain.LetterStateCorrect,
				domain.LetterStateCorrect,
				domain.LetterStateCorrect,
			},
		},
		{
			name:   "Частичное совпадение",
			guess:  "pleap",
			target: "apple",
			want: []domain.LetterState{
				domain.LetterStatePresent,
				domain.LetterStatePresent,
				domain.LetterStatePresent,
				domain.LetterStatePresent,
				domain.LetterStatePresent,
			},
		},
		{
			name:   "Дубликаты букв с частичным совпадением",
			guess:  "paper",
			target: "apple",
			want: []domain.LetterState{
				domain.LetterStatePresent,
				domain.LetterStatePresent,
				domain.LetterStateCorrect,
				domain.LetterStatePresent,
				domain.LetterStateAbsent,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluateGuess(tt.guess, tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWordleUseCase_GetDailyState(t *testing.T) {
	ctx := context.Background()
	userID := int64(42)

	tests := []struct {
		name         string
		mockBehavior func(m *mocks.MockWordleRepository)
		want         domain.DailyGameState
		wantErr      error
	}{
		{
			name: "Ошибка получения слова дня",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("", errors.New("db error"))
			},
			want:    domain.DailyGameState{},
			wantErr: errutil.Internal("failed to get daily word", errors.New("db error")),
		},
		{
			name: "Ошибка получения состояния игры",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("apple", nil)
				m.EXPECT().GetGameState(gomock.Any(), userID, gomock.Any()).Return(domain.WordleGame{}, nil, errors.New("db error"))
			},
			want:    domain.DailyGameState{},
			wantErr: errutil.Internal("failed to get game state", errors.New("db error")),
		},
		{
			name: "Успешное получение состояния в процессе",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("apple", nil)
				m.EXPECT().GetGameState(gomock.Any(), userID, gomock.Any()).Return(
					domain.WordleGame{Solved: false, Attempt: 2},
					[]domain.WordleGuess{{Word: "arise"}},
					nil,
				)
				m.EXPECT().GetCurrentStreak(gomock.Any(), userID).Return(int32(5), nil)
			},
			want: domain.DailyGameState{
				Status:        domain.GameStatusPlaying,
				CurrentStreak: 5,
				Guesses: []domain.WordleGuessResult{
					{
						Word: "arise",
						Letters: []domain.LetterState{
							domain.LetterStateCorrect,
							domain.LetterStateAbsent,
							domain.LetterStateAbsent,
							domain.LetterStateAbsent,
							domain.LetterStateCorrect,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Успешное получение состояния победа с ошибкой получения стрика",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("apple", nil)
				m.EXPECT().GetGameState(gomock.Any(), userID, gomock.Any()).Return(
					domain.WordleGame{Solved: true, Attempt: 1},
					nil,
					nil,
				)
				m.EXPECT().GetCurrentStreak(gomock.Any(), userID).Return(int32(0), errors.New("streak error"))
			},
			want: domain.DailyGameState{
				Status:     domain.GameStatusWon,
				TargetWord: "apple",
			},
			wantErr: nil,
		},
		{
			name: "Успешное получение состояния поражение",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("apple", nil)
				m.EXPECT().GetGameState(gomock.Any(), userID, gomock.Any()).Return(
					domain.WordleGame{Solved: false, Attempt: domain.WordleMaxAttempts},
					nil,
					nil,
				)
				m.EXPECT().GetCurrentStreak(gomock.Any(), userID).Return(int32(2), nil)
			},
			want: domain.DailyGameState{
				Status:        domain.GameStatusLost,
				TargetWord:    "apple",
				CurrentStreak: 2,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockWordleRepository(ctrl)
			tt.mockBehavior(repo)

			uc := NewWordleUseCase(repo, nil)
			got, err := uc.GetDailyState(ctx, userID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestWordleUseCase_MakeGuess(t *testing.T) {
	ctx := context.Background()
	userID := int64(10)
	idemKey := "key123"
	now := time.Now()

	tests := []struct {
		name         string
		guess        string
		mockBehavior func(m *mocks.MockWordleRepository)
		want         domain.MakeWordleGuessResult
		wantErr      error
	}{
		{
			name:  "Неверная длина слова",
			guess: "app",
			mockBehavior: func(m *mocks.MockWordleRepository) {
			},
			want:    domain.MakeWordleGuessResult{},
			wantErr: domain.ErrInvalidWordLength,
		},
		{
			name:  "Ошибка получения слова дня",
			guess: "apple",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("", errors.New("word error"))
			},
			want:    domain.MakeWordleGuessResult{},
			wantErr: errutil.Internal("failed to get daily word", errors.New("word error")),
		},
		{
			name:  "Ошибка получения состояния игры",
			guess: "apple",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("apple", nil)
				m.EXPECT().GetGameState(gomock.Any(), userID, gomock.Any()).Return(domain.WordleGame{}, nil, errors.New("state error"))
			},
			want:    domain.MakeWordleGuessResult{},
			wantErr: errutil.Internal("failed to get game state", errors.New("state error")),
		},
		{
			name:  "Игра уже завершена",
			guess: "apple",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("apple", nil)
				m.EXPECT().GetGameState(gomock.Any(), userID, gomock.Any()).Return(
					domain.WordleGame{FinishedAt: &now},
					nil,
					nil,
				)
			},
			want:    domain.MakeWordleGuessResult{},
			wantErr: domain.ErrGameAlreadyFinished,
		},
		{
			name:  "Ошибка проверки слова в словаре",
			guess: "apple",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("apple", nil)
				m.EXPECT().GetGameState(gomock.Any(), userID, gomock.Any()).Return(domain.WordleGame{}, nil, nil)
				m.EXPECT().CheckWordExists(gomock.Any(), "apple").Return(false, errors.New("dict error"))
			},
			want:    domain.MakeWordleGuessResult{},
			wantErr: errutil.Internal("failed to check word existence", errors.New("dict error")),
		},
		{
			name:  "Слова нет в словаре",
			guess: "xxxxx",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("apple", nil)
				m.EXPECT().GetGameState(gomock.Any(), userID, gomock.Any()).Return(domain.WordleGame{}, nil, nil)
				m.EXPECT().CheckWordExists(gomock.Any(), "xxxxx").Return(false, nil)
			},
			want:    domain.MakeWordleGuessResult{},
			wantErr: domain.ErrWordNotInDictionary,
		},
		{
			name:  "Конфликт идемпотентности",
			guess: "arise",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("apple", nil)
				m.EXPECT().GetGameState(gomock.Any(), userID, gomock.Any()).Return(domain.WordleGame{Attempt: 1}, nil, nil)
				m.EXPECT().CheckWordExists(gomock.Any(), "arise").Return(true, nil)
				m.EXPECT().SaveGuessWithTransaction(gomock.Any(), gomock.Any(), false, false).Return(domain.ErrIdempotencyConflict)
			},
			want:    domain.MakeWordleGuessResult{},
			wantErr: domain.ErrIdempotencyConflict,
		},
		{
			name:  "Внутренняя ошибка сохранения попытки",
			guess: "arise",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("apple", nil)
				m.EXPECT().GetGameState(gomock.Any(), userID, gomock.Any()).Return(domain.WordleGame{Attempt: 1}, nil, nil)
				m.EXPECT().CheckWordExists(gomock.Any(), "arise").Return(true, nil)
				m.EXPECT().SaveGuessWithTransaction(gomock.Any(), gomock.Any(), false, false).Return(errors.New("tx error"))
			},
			want:    domain.MakeWordleGuessResult{},
			wantErr: errutil.Internal("failed to save guess", errors.New("tx error")),
		},
		{
			name:  "Успешная попытка игра продолжается",
			guess: "arise",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("apple", nil)
				m.EXPECT().GetGameState(gomock.Any(), userID, gomock.Any()).Return(domain.WordleGame{Attempt: 1}, nil, nil)
				m.EXPECT().CheckWordExists(gomock.Any(), "arise").Return(true, nil)
				m.EXPECT().SaveGuessWithTransaction(gomock.Any(), gomock.Any(), false, false).Return(nil)
				m.EXPECT().GetCurrentStreak(gomock.Any(), userID).Return(int32(3), nil)
				m.EXPECT().CountWins(gomock.Any(), userID).Return(int32(10), nil)
			},
			want: domain.MakeWordleGuessResult{
				Status: domain.GameStatusPlaying,
				GuessResult: domain.WordleGuessResult{
					Word: "arise",
					Letters: []domain.LetterState{
						domain.LetterStateCorrect,
						domain.LetterStateAbsent,
						domain.LetterStateAbsent,
						domain.LetterStateAbsent,
						domain.LetterStateCorrect,
					},
				},
				CurrentStreak: 3,
				TotalWins:     10,
			},
			wantErr: nil,
		},
		{
			name:  "Успешная попытка победа",
			guess: "apple",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("apple", nil)
				m.EXPECT().GetGameState(gomock.Any(), userID, gomock.Any()).Return(domain.WordleGame{Attempt: 1}, nil, nil)
				m.EXPECT().CheckWordExists(gomock.Any(), "apple").Return(true, nil)
				m.EXPECT().SaveGuessWithTransaction(gomock.Any(), gomock.Any(), true, false).Return(nil)
				m.EXPECT().GetCurrentStreak(gomock.Any(), userID).Return(int32(4), nil)
				m.EXPECT().CountWins(gomock.Any(), userID).Return(int32(11), nil)
			},
			want: domain.MakeWordleGuessResult{
				Status: domain.GameStatusWon,
				GuessResult: domain.WordleGuessResult{
					Word: "apple",
					Letters: []domain.LetterState{
						domain.LetterStateCorrect,
						domain.LetterStateCorrect,
						domain.LetterStateCorrect,
						domain.LetterStateCorrect,
						domain.LetterStateCorrect,
					},
				},
				TargetWord:    "apple",
				CurrentStreak: 4,
				TotalWins:     11,
			},
			wantErr: nil,
		},
		{
			name:  "Успешная попытка поражение",
			guess: "arise",
			mockBehavior: func(m *mocks.MockWordleRepository) {
				m.EXPECT().GetDailyWord(gomock.Any(), gomock.Any()).Return("apple", nil)
				m.EXPECT().GetGameState(gomock.Any(), userID, gomock.Any()).Return(domain.WordleGame{Attempt: domain.WordleMaxAttempts - 1}, nil, nil)
				m.EXPECT().CheckWordExists(gomock.Any(), "arise").Return(true, nil)
				m.EXPECT().SaveGuessWithTransaction(gomock.Any(), gomock.Any(), false, true).Return(nil)
				m.EXPECT().GetCurrentStreak(gomock.Any(), userID).Return(int32(0), errors.New("streak err"))
				m.EXPECT().CountWins(gomock.Any(), userID).Return(int32(0), errors.New("wins err"))
			},
			want: domain.MakeWordleGuessResult{
				Status: domain.GameStatusLost,
				GuessResult: domain.WordleGuessResult{
					Word: "arise",
					Letters: []domain.LetterState{
						domain.LetterStateCorrect,
						domain.LetterStateAbsent,
						domain.LetterStateAbsent,
						domain.LetterStateAbsent,
						domain.LetterStateCorrect,
					},
				},
				TargetWord: "apple",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockWordleRepository(ctrl)
			tt.mockBehavior(repo)

			uc := NewWordleUseCase(repo, nil)
			got, err := uc.MakeGuess(ctx, userID, tt.guess, idemKey)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
