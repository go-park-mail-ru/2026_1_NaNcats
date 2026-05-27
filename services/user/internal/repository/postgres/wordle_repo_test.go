package postgres

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWordleRepo_GetDailyWord(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC)
	dateStr := "2026-05-27"

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   string
		expectedError error
	}{
		{
			name: "Успешное получение слова дня",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT w.word FROM "wordle_daily"`)).
					WithArgs(dateStr).
					WillReturnRows(pgxmock.NewRows([]string{"word"}).AddRow("APPLE"))
			},
			expectedRes:   "APPLE",
			expectedError: nil,
		},
		{
			name: "Слово дня не найдено",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT w.word FROM "wordle_daily"`)).
					WithArgs(dateStr).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedRes:   "",
			expectedError: domain.ErrDailyWordNotFound,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT w.word FROM "wordle_daily"`)).
					WithArgs(dateStr).
					WillReturnError(errors.New("db error"))
			},
			expectedRes:   "",
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewWordleRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetDailyWord(ctx, date)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestWordleRepo_CheckWordExists(t *testing.T) {
	ctx := context.Background()
	word := "HELLO"

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   bool
		expectedError error
	}{
		{
			name: "Слово существует",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM "wordle_word"`)).
					WithArgs(word).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
			},
			expectedRes:   true,
			expectedError: nil,
		},
		{
			name: "Слово не существует",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM "wordle_word"`)).
					WithArgs(word).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))
			},
			expectedRes:   false,
			expectedError: nil,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM "wordle_word"`)).
					WithArgs(word).
					WillReturnError(errors.New("db error"))
			},
			expectedRes:   false,
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewWordleRepo(mock)
			tt.mockInit(mock)

			res, err := repo.CheckWordExists(ctx, word)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestWordleRepo_GetGameState(t *testing.T) {
	ctx := context.Background()
	var userID int64 = 42
	date := time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC)
	dateStr := "2026-05-27"
	now := time.Now()

	gameCols := []string{"user_id", "game_date", "solved", "attempt", "finished_at", "created_at", "updated_at"}
	guessCols := []string{"user_id", "guess_date", "attempt_num", "word", "idempotency_key", "created_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedGame  domain.WordleGame
		expectedArr   []domain.WordleGuess
		expectedError error
	}{
		{
			name: "Успешное получение игры и попыток",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, game_date, solved, attempt, finished_at, created_at, updated_at`)).
					WithArgs(userID, dateStr).
					WillReturnRows(pgxmock.NewRows(gameCols).AddRow(userID, date, true, int32(2), &now, now, now))

				m.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, guess_date, attempt_num, word, idempotency_key, created_at`)).
					WithArgs(userID, dateStr).
					WillReturnRows(pgxmock.NewRows(guessCols).
						AddRow(userID, date, int32(1), "AUDIO", "key1", now).
						AddRow(userID, date, int32(2), "APPLE", "key2", now))
			},
			expectedGame: domain.WordleGame{
				UserID:     userID,
				GameDate:   date,
				Solved:     true,
				Attempt:    2,
				FinishedAt: &now,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			expectedArr: []domain.WordleGuess{
				{UserID: userID, GuessDate: date, AttemptNum: 1, Word: "AUDIO", IdempotencyKey: "key1", CreatedAt: now},
				{UserID: userID, GuessDate: date, AttemptNum: 2, Word: "APPLE", IdempotencyKey: "key2", CreatedAt: now},
			},
			expectedError: nil,
		},
		{
			name: "Игра не найдена (возврат дефолтной игры)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, game_date`)).
					WithArgs(userID, dateStr).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedGame: domain.WordleGame{
				UserID:   userID,
				GameDate: date,
				Solved:   false,
				Attempt:  0,
			},
			expectedArr:   nil,
			expectedError: nil,
		},
		{
			name: "Ошибка базы данных при получении игры",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, game_date`)).
					WithArgs(userID, dateStr).
					WillReturnError(errors.New("db game error"))
			},
			expectedGame:  domain.WordleGame{},
			expectedArr:   nil,
			expectedError: errors.New("db game error"),
		},
		{
			name: "Ошибка базы данных при получении попыток",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, game_date`)).
					WithArgs(userID, dateStr).
					WillReturnRows(pgxmock.NewRows(gameCols).AddRow(userID, date, false, int32(1), nil, now, now))

				m.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, guess_date`)).
					WithArgs(userID, dateStr).
					WillReturnError(errors.New("db guess error"))
			},
			expectedGame:  domain.WordleGame{},
			expectedArr:   nil,
			expectedError: errors.New("db guess error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewWordleRepo(mock)
			tt.mockInit(mock)

			game, guesses, err := repo.GetGameState(ctx, userID, date)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedGame, game)
				assert.Equal(t, tt.expectedArr, guesses)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestWordleRepo_GetCurrentStreak(t *testing.T) {
	ctx := context.Background()
	var userID int64 = 42

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   int32
		expectedError error
	}{
		{
			name: "Стрик найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT current_streak FROM "wordle_streak"`)).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows([]string{"current_streak"}).AddRow(int32(5)))
			},
			expectedRes:   5,
			expectedError: nil,
		},
		{
			name: "Стрик не найден (возврат 0)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT current_streak FROM "wordle_streak"`)).
					WithArgs(userID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedRes:   0,
			expectedError: nil,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT current_streak FROM "wordle_streak"`)).
					WithArgs(userID).
					WillReturnError(errors.New("db error"))
			},
			expectedRes:   0,
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewWordleRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetCurrentStreak(ctx, userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestWordleRepo_CountWins(t *testing.T) {
	ctx := context.Background()
	var userID int64 = 42

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   int32
		expectedError error
	}{
		{
			name: "Успешный подсчет",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "wordle_game"`)).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(10)))
			},
			expectedRes:   10,
			expectedError: nil,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "wordle_game"`)).
					WithArgs(userID).
					WillReturnError(errors.New("db error"))
			},
			expectedRes:   0,
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewWordleRepo(mock)
			tt.mockInit(mock)

			res, err := repo.CountWins(ctx, userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
