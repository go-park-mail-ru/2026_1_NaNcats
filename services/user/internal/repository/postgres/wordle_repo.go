package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type wordleRepo struct {
	pool postgres.PgxPool
}

func NewWordleRepo(pool postgres.PgxPool) repository.WordleRepository {
	return &wordleRepo{
		pool: pool,
	}
}

func (r *wordleRepo) GetDailyWord(ctx context.Context, date time.Time) (string, error) {
	query := `
		SELECT w.word 
		FROM "wordle_daily" wd
		JOIN "wordle_word" w ON wd.word_id = w.id
		WHERE wd.word_of_day = $1
	`
	var word string
	err := r.pool.QueryRow(ctx, query, date.Format("2006-01-02")).Scan(&word)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrDailyWordNotFound
		}
		return "", err
	}
	return word, nil
}

func (r *wordleRepo) CheckWordExists(ctx context.Context, word string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM "wordle_word" WHERE word = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, word).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *wordleRepo) GetGameState(ctx context.Context, userID int64, date time.Time) (domain.WordleGame, []domain.WordleGuess, error) {
	gameQuery := `
		SELECT user_id, game_date, solved, attempt, finished_at, created_at, updated_at
		FROM "wordle_game"
		WHERE user_id = $1 AND game_date = $2
	`
	var game domain.WordleGame
	dateStr := date.Format("2006-01-02")

	err := r.pool.QueryRow(ctx, gameQuery, userID, dateStr).Scan(
		&game.UserID,
		&game.GameDate,
		&game.Solved,
		&game.Attempt,
		&game.FinishedAt,
		&game.CreatedAt,
		&game.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			game = domain.WordleGame{
				UserID:   userID,
				GameDate: date,
				Solved:   false,
				Attempt:  0,
			}
			return game, nil, nil
		}
		return domain.WordleGame{}, nil, err
	}

	guessesQuery := `
		SELECT user_id, guess_date, attempt_num, word, idempotency_key, created_at
		FROM "wordle_guess"
		WHERE user_id = $1 AND guess_date = $2
		ORDER BY attempt_num ASC
	`
	rows, err := r.pool.Query(ctx, guessesQuery, userID, dateStr)
	if err != nil {
		return domain.WordleGame{}, nil, err
	}
	defer rows.Close()

	var guesses []domain.WordleGuess
	for rows.Next() {
		var g domain.WordleGuess
		err := rows.Scan(&g.UserID, &g.GuessDate, &g.AttemptNum, &g.Word, &g.IdempotencyKey, &g.CreatedAt)
		if err != nil {
			return domain.WordleGame{}, nil, err
		}
		guesses = append(guesses, g)
	}

	return game, guesses, rows.Err()
}

func (r *wordleRepo) GetCurrentStreak(ctx context.Context, userID int64) (int32, error) {
	var streak int32
	err := r.pool.QueryRow(ctx,
		`SELECT current_streak FROM "wordle_streak" WHERE user_id = $1`,
		userID,
	).Scan(&streak)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return streak, nil
}

func (r *wordleRepo) CountWins(ctx context.Context, userID int64) (int32, error) {
	var count int32
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM "wordle_game" WHERE user_id = $1 AND solved = true`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *wordleRepo) SaveGuessWithTransaction(ctx context.Context, guess domain.WordleGuess, isWin, isLoss bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	dateStr := guess.GuessDate.Format("2006-01-02")

	var finishedAt *time.Time
	if isWin || isLoss {
		now := time.Now()
		finishedAt = &now
	}

	// wordle_guess ссылается на wordle_game по (user_id, game_date), поэтому
	// сначала апсертим саму игру, потом сохраняем попытку.
	upsertGameQuery := `
		INSERT INTO "wordle_game" (user_id, game_date, solved, attempt, finished_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, game_date) DO UPDATE
		SET solved = EXCLUDED.solved,
		    attempt = EXCLUDED.attempt,
		    finished_at = EXCLUDED.finished_at,
		    updated_at = NOW()
	`
	_, err = tx.Exec(ctx, upsertGameQuery, guess.UserID, dateStr, isWin, guess.AttemptNum, finishedAt)
	if err != nil {
		return err
	}

	insertGuessQuery := `
		INSERT INTO "wordle_guess" (user_id, guess_date, attempt_num, word, idempotency_key)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (idempotency_key) DO NOTHING
	`
	tag, err := tx.Exec(ctx, insertGuessQuery, guess.UserID, dateStr, guess.AttemptNum, guess.Word, guess.IdempotencyKey)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrIdempotencyConflict
	}

	if isWin {
		updateStreakQuery := `
			INSERT INTO "wordle_streak" (user_id, current_streak, last_played)
			VALUES ($1, 1, $2)
			ON CONFLICT (user_id) DO UPDATE 
			SET current_streak = CASE 
			        WHEN "wordle_streak".last_played = EXCLUDED.last_played - 1 THEN "wordle_streak".current_streak + 1
			        WHEN "wordle_streak".last_played = EXCLUDED.last_played THEN "wordle_streak".current_streak
			        ELSE 1 
			    END,
			    last_played = EXCLUDED.last_played,
			    updated_at = NOW()
		`
		_, err = tx.Exec(ctx, updateStreakQuery, guess.UserID, dateStr)
		if err != nil {
			return err
		}
	}

	if isLoss {
		updateStreakLossQuery := `
			INSERT INTO "wordle_streak" (user_id, current_streak, last_played)
			VALUES ($1, 0, $2)
			ON CONFLICT (user_id) DO UPDATE 
			SET current_streak = 0, 
			    last_played = EXCLUDED.last_played, 
			    updated_at = NOW()
		`
		_, err = tx.Exec(ctx, updateStreakLossQuery, guess.UserID, dateStr)
		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
