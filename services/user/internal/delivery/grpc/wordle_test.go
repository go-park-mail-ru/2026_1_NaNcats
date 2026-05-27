package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase/mocks"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGameHandler_GetDailyWordleState(t *testing.T) {
	type mockInit func(w *mocks.MockWordleUseCase)
	req := &pb.GetDailyWordleStateRequest{UserId: 42}

	domainState := domain.DailyGameState{
		Status:        domain.GameStatusPlaying,
		CurrentStreak: 5,
		Guesses: []domain.WordleGuessResult{
			{
				Word: "APPLE",
				Letters: []domain.LetterState{
					domain.LetterStateCorrect,
					domain.LetterStatePresent,
					domain.LetterStateAbsent,
					domain.LetterStateCorrect,
					domain.LetterStateAbsent,
				},
			},
		},
	}

	pbGuesses := []*pb.WordleGuessResult{
		{
			Word: "APPLE",
			Letters: []pb.LetterState{
				pb.LetterState_LETTER_STATE_CORRECT,
				pb.LetterState_LETTER_STATE_PRESENT,
				pb.LetterState_LETTER_STATE_ABSENT,
				pb.LetterState_LETTER_STATE_CORRECT,
				pb.LetterState_LETTER_STATE_ABSENT,
			},
		},
	}

	tests := []struct {
		name         string
		req          *pb.GetDailyWordleStateRequest
		mockInit     mockInit
		expectedCode codes.Code
		expectedResp *pb.GetDailyWordleStateResponse
	}{
		{
			name: "Успешное получение состояния игры",
			req:  req,
			mockInit: func(w *mocks.MockWordleUseCase) {
				w.EXPECT().GetDailyState(gomock.Any(), int64(42)).Return(domainState, nil)
			},
			expectedCode: codes.OK,
			expectedResp: &pb.GetDailyWordleStateResponse{
				WordLength:    domain.WordleWordLength,
				MaxAttempts:   domain.WordleMaxAttempts,
				Status:        pb.GameStatus_GAME_STATUS_PLAYING,
				CurrentStreak: 5,
				Guesses:       pbGuesses,
				TargetWord:    "",
			},
		},
		{
			name: "Состояние игры: Выиграно",
			req:  req,
			mockInit: func(w *mocks.MockWordleUseCase) {
				wonState := domainState
				wonState.Status = domain.GameStatusWon
				wonState.TargetWord = "APPLE"
				w.EXPECT().GetDailyState(gomock.Any(), int64(42)).Return(wonState, nil)
			},
			expectedCode: codes.OK,
			expectedResp: &pb.GetDailyWordleStateResponse{
				WordLength:    domain.WordleWordLength,
				MaxAttempts:   domain.WordleMaxAttempts,
				Status:        pb.GameStatus_GAME_STATUS_WON,
				CurrentStreak: 5,
				Guesses:       pbGuesses,
				TargetWord:    "APPLE",
			},
		},
		{
			name: "Состояние игры: Проиграно",
			req:  req,
			mockInit: func(w *mocks.MockWordleUseCase) {
				lostState := domainState
				lostState.Status = domain.GameStatusLost
				lostState.TargetWord = "AUDIO"
				w.EXPECT().GetDailyState(gomock.Any(), int64(42)).Return(lostState, nil)
			},
			expectedCode: codes.OK,
			expectedResp: &pb.GetDailyWordleStateResponse{
				WordLength:    domain.WordleWordLength,
				MaxAttempts:   domain.WordleMaxAttempts,
				Status:        pb.GameStatus_GAME_STATUS_LOST,
				CurrentStreak: 5,
				Guesses:       pbGuesses,
				TargetWord:    "AUDIO",
			},
		},
		{
			name: "Неизвестный статус игры",
			req:  req,
			mockInit: func(w *mocks.MockWordleUseCase) {
				unknownState := domainState
				unknownState.Status = "UNKNOWN"
				w.EXPECT().GetDailyState(gomock.Any(), int64(42)).Return(unknownState, nil)
			},
			expectedCode: codes.OK,
			expectedResp: &pb.GetDailyWordleStateResponse{
				WordLength:    domain.WordleWordLength,
				MaxAttempts:   domain.WordleMaxAttempts,
				Status:        pb.GameStatus_GAME_STATUS_UNSPECIFIED,
				CurrentStreak: 5,
				Guesses:       pbGuesses,
				TargetWord:    "",
			},
		},
		{
			name: "Неизвестный статус буквы",
			req:  req,
			mockInit: func(w *mocks.MockWordleUseCase) {
				unknownLetterState := domainState
				unknownLetterState.Guesses[0].Letters[0] = "UNKNOWN"
				w.EXPECT().GetDailyState(gomock.Any(), int64(42)).Return(unknownLetterState, nil)
			},
			expectedCode: codes.OK,
			expectedResp: &pb.GetDailyWordleStateResponse{
				WordLength:    domain.WordleWordLength,
				MaxAttempts:   domain.WordleMaxAttempts,
				Status:        pb.GameStatus_GAME_STATUS_PLAYING,
				CurrentStreak: 5,
				Guesses: []*pb.WordleGuessResult{
					{
						Word: "APPLE",
						Letters: []pb.LetterState{
							pb.LetterState_LETTER_STATE_UNSPECIFIED,
							pb.LetterState_LETTER_STATE_PRESENT,
							pb.LetterState_LETTER_STATE_ABSENT,
							pb.LetterState_LETTER_STATE_CORRECT,
							pb.LetterState_LETTER_STATE_ABSENT,
						},
					},
				},
				TargetWord: "",
			},
		},
		{
			name: "Ошибка при получении состояния игры",
			req:  req,
			mockInit: func(w *mocks.MockWordleUseCase) {
				w.EXPECT().GetDailyState(gomock.Any(), int64(42)).Return(domain.DailyGameState{}, errors.New("db error"))
			},
			expectedCode: codes.Internal,
			expectedResp: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			wordleUC := mocks.NewMockWordleUseCase(ctrl)
			tt.mockInit(wordleUC)

			h := NewGameHandler(wordleUC)
			resp, err := h.GetDailyWordleState(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestGameHandler_MakeWordleGuess(t *testing.T) {
	type mockInit func(w *mocks.MockWordleUseCase)
	req := &pb.MakeWordleGuessRequest{
		UserId:         42,
		Guess:          "APPLE",
		IdempotencyKey: "idem-key",
	}

	domainResult := domain.MakeWordleGuessResult{
		Status: domain.GameStatusPlaying,
		GuessResult: domain.WordleGuessResult{
			Word: "APPLE",
			Letters: []domain.LetterState{
				domain.LetterStateCorrect,
				domain.LetterStatePresent,
				domain.LetterStateAbsent,
				domain.LetterStateCorrect,
				domain.LetterStateAbsent,
			},
		},
		CurrentStreak: 5,
		TotalWins:     10,
	}

	pbGuessResult := &pb.WordleGuessResult{
		Word: "APPLE",
		Letters: []pb.LetterState{
			pb.LetterState_LETTER_STATE_CORRECT,
			pb.LetterState_LETTER_STATE_PRESENT,
			pb.LetterState_LETTER_STATE_ABSENT,
			pb.LetterState_LETTER_STATE_CORRECT,
			pb.LetterState_LETTER_STATE_ABSENT,
		},
	}

	tests := []struct {
		name         string
		req          *pb.MakeWordleGuessRequest
		mockInit     mockInit
		expectedCode codes.Code
		expectedResp *pb.MakeWordleGuessResponse
	}{
		{
			name: "Успешная попытка (продолжение игры)",
			req:  req,
			mockInit: func(w *mocks.MockWordleUseCase) {
				w.EXPECT().MakeGuess(gomock.Any(), int64(42), "APPLE", "idem-key").Return(domainResult, nil)
			},
			expectedCode: codes.OK,
			expectedResp: &pb.MakeWordleGuessResponse{
				Status:        pb.GameStatus_GAME_STATUS_PLAYING,
				GuessResult:   pbGuessResult,
				CurrentStreak: 5,
				TotalWins:     10,
				TargetWord:    "",
			},
		},
		{
			name: "Успешная попытка (победа)",
			req:  req,
			mockInit: func(w *mocks.MockWordleUseCase) {
				wonResult := domainResult
				wonResult.Status = domain.GameStatusWon
				wonResult.TargetWord = "APPLE"
				w.EXPECT().MakeGuess(gomock.Any(), int64(42), "APPLE", "idem-key").Return(wonResult, nil)
			},
			expectedCode: codes.OK,
			expectedResp: &pb.MakeWordleGuessResponse{
				Status:        pb.GameStatus_GAME_STATUS_WON,
				GuessResult:   pbGuessResult,
				CurrentStreak: 5,
				TotalWins:     10,
				TargetWord:    "APPLE",
			},
		},
		{
			name: "Успешная попытка (поражение)",
			req:  req,
			mockInit: func(w *mocks.MockWordleUseCase) {
				lostResult := domainResult
				lostResult.Status = domain.GameStatusLost
				lostResult.TargetWord = "AUDIO"
				w.EXPECT().MakeGuess(gomock.Any(), int64(42), "APPLE", "idem-key").Return(lostResult, nil)
			},
			expectedCode: codes.OK,
			expectedResp: &pb.MakeWordleGuessResponse{
				Status:        pb.GameStatus_GAME_STATUS_LOST,
				GuessResult:   pbGuessResult,
				CurrentStreak: 5,
				TotalWins:     10,
				TargetWord:    "AUDIO",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			wordleUC := mocks.NewMockWordleUseCase(ctrl)
			tt.mockInit(wordleUC)

			h := NewGameHandler(wordleUC)
			resp, err := h.MakeWordleGuess(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}
