package gameclient

import (
	"context"
	"errors"
	"strings"

	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrGameAlreadyFinished = errors.New("game already finished for today")
	ErrWordNotInDictionary = errors.New("word is not in the dictionary")
	ErrInvalidWordLength   = errors.New("invalid word length")
	ErrIdempotencyConflict = errors.New("idempotency conflict: request already processed")
	ErrInternal            = errors.New("internal server error")
)

//go:generate mockgen -destination=mocks/game_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/gameclient GameClient
type GameClient interface {
	GetDailyWordleState(ctx context.Context, userID int64) (*pb.GetDailyWordleStateResponse, error)
	MakeWordleGuess(ctx context.Context, userID int64, guess, idempotencyKey string) (*pb.MakeWordleGuessResponse, error)
}

type gameClient struct {
	client pb.WordleServiceClient
}

func NewGameClient(cl pb.WordleServiceClient) GameClient {
	return &gameClient{client: cl}
}

func (c *gameClient) GetDailyWordleState(ctx context.Context, userID int64) (*pb.GetDailyWordleStateResponse, error) {
	resp, err := c.client.GetDailyWordleState(ctx, &pb.GetDailyWordleStateRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, ErrInternal
	}
	return resp, nil
}

func (c *gameClient) MakeWordleGuess(ctx context.Context, userID int64, guess, idempotencyKey string) (*pb.MakeWordleGuessResponse, error) {
	resp, err := c.client.MakeWordleGuess(ctx, &pb.MakeWordleGuessRequest{
		UserId:         userID,
		Guess:          guess,
		IdempotencyKey: idempotencyKey,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.AlreadyExists:
				return nil, ErrIdempotencyConflict
			case codes.FailedPrecondition:
				return nil, ErrGameAlreadyFinished
			case codes.InvalidArgument:
				// Сообщение от gRPC может прийти как человекочитаемый текст
				// ("word is not in the dictionary"), так и как слаг
				// ("WORD_NOT_IN_DICTIONARY") — матчим без учёта регистра, иначе
				// ошибка «нет в словаре» ошибочно классифицировалась как длина.
				if strings.Contains(strings.ToLower(st.Message()), "dictionary") {
					return nil, ErrWordNotInDictionary
				}
				return nil, ErrInvalidWordLength
			}
		}
		return nil, ErrInternal
	}

	return resp, nil
}
