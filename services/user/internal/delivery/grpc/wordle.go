package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
)

type GameHandler struct {
	pb.UnimplementedWordleServiceServer
	wordleUC usecase.WordleUseCase
}

func NewGameHandler(wuc usecase.WordleUseCase) *GameHandler {
	return &GameHandler{
		wordleUC: wuc,
	}
}

func mapDomainStatusToPB(status domain.GameStatus) pb.GameStatus {
	switch status {
	case domain.GameStatusPlaying:
		return pb.GameStatus_GAME_STATUS_PLAYING
	case domain.GameStatusWon:
		return pb.GameStatus_GAME_STATUS_WON
	case domain.GameStatusLost:
		return pb.GameStatus_GAME_STATUS_LOST
	default:
		return pb.GameStatus_GAME_STATUS_UNSPECIFIED
	}
}

func mapDomainLettersToPB(letters []domain.LetterState) []pb.LetterState {
	pbLetters := make([]pb.LetterState, len(letters))
	for i, l := range letters {
		switch l {
		case domain.LetterStateCorrect:
			pbLetters[i] = pb.LetterState_LETTER_STATE_CORRECT
		case domain.LetterStatePresent:
			pbLetters[i] = pb.LetterState_LETTER_STATE_PRESENT
		case domain.LetterStateAbsent:
			pbLetters[i] = pb.LetterState_LETTER_STATE_ABSENT
		default:
			pbLetters[i] = pb.LetterState_LETTER_STATE_UNSPECIFIED
		}
	}
	return pbLetters
}

func (h *GameHandler) GetDailyWordleState(ctx context.Context, req *pb.GetDailyWordleStateRequest) (*pb.GetDailyWordleStateResponse, error) {
	state, err := h.wordleUC.GetDailyState(ctx, req.UserId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	var pbGuesses []*pb.WordleGuessResult
	for _, g := range state.Guesses {
		pbGuesses = append(pbGuesses, &pb.WordleGuessResult{
			Word:    g.Word,
			Letters: mapDomainLettersToPB(g.Letters),
		})
	}

	return &pb.GetDailyWordleStateResponse{
		WordLength:  domain.WordleWordLength,
		MaxAttempts: domain.WordleMaxAttempts,
		Status:      mapDomainStatusToPB(state.Status),
		Guesses:     pbGuesses,
	}, nil
}

func (h *GameHandler) MakeWordleGuess(ctx context.Context, req *pb.MakeWordleGuessRequest) (*pb.MakeWordleGuessResponse, error) {
	result, err := h.wordleUC.MakeGuess(ctx, req.UserId, req.Guess, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.MakeWordleGuessResponse{
		Status: mapDomainStatusToPB(result.Status),
		GuessResult: &pb.WordleGuessResult{
			Word:    result.GuessResult.Word,
			Letters: mapDomainLettersToPB(result.GuessResult.Letters),
		},
		BonusAwarded: result.BonusAwarded,
	}, nil
}
