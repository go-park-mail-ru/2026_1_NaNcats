package domain

import "time"

const (
	WordleWordLength  = 5
	WordleMaxAttempts = 6
	WordleWinBonus    = 500
)

type GameStatus string

const (
	GameStatusPlaying GameStatus = "PLAYING"
	GameStatusWon     GameStatus = "WON"
	GameStatusLost    GameStatus = "LOST"
)

type LetterState string

const (
	LetterStateCorrect LetterState = "CORRECT"
	LetterStatePresent LetterState = "PRESENT"
	LetterStateAbsent  LetterState = "ABSENT"
)

type WordleGame struct {
	UserID     int64
	GameDate   time.Time
	Solved     bool
	Attempt    int
	FinishedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type WordleGuess struct {
	UserID         int64
	GuessDate      time.Time
	AttemptNum     int
	Word           string
	IdempotencyKey string
	CreatedAt      time.Time
}

type WordleGuessResult struct {
	Word    string
	Letters []LetterState
}

type DailyGameState struct {
	Status  GameStatus
	Guesses []WordleGuessResult
}

type MakeWordleGuessResult struct {
	Status       GameStatus
	GuessResult  WordleGuessResult
	BonusAwarded int64
}
