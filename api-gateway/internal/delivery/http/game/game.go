package game

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/gameclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
	pbOrder "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
	"github.com/microcosm-cc/bluemonday"
)

// WordleWinPromoDiscount — фиксированный размер промо за победу (в микрорублях).
const WordleWinPromoDiscount = int64(100_000_000)

//go:generate easyjson $GOFILE

//easyjson:json
type MakeWordleGuessRequest struct {
	Guess string `json:"guess" example:"apple"`
}

func (r *MakeWordleGuessRequest) Sanitize(p *bluemonday.Policy) {
	r.Guess = p.Sanitize(r.Guess)
}

//easyjson:json
type WordleGuessResultDTO struct {
	Word    string   `json:"word" example:"apple"`
	Letters []string `json:"letters" example:"['CORRECT', 'PRESENT', 'ABSENT', 'ABSENT', 'ABSENT']"`
}

//easyjson:json
type DailyStateResponse struct {
	WordLength    int32                  `json:"word_length" example:"5"`
	MaxAttempts   int32                  `json:"max_attempts" example:"6"`
	Status        string                 `json:"status" example:"PLAYING"` // PLAYING, WON, LOST
	Guesses       []WordleGuessResultDTO `json:"guesses"`
	CurrentStreak int32                  `json:"current_streak" example:"3"`
	TargetWord    string                 `json:"target_word,omitempty" example:"пицца"`
}

//easyjson:json
type MakeGuessResponse struct {
	Status        string               `json:"status" example:"WON"`
	GuessResult   WordleGuessResultDTO `json:"guess_result"`
	CurrentStreak int32                `json:"current_streak" example:"3"`
	TargetWord    string               `json:"target_word,omitempty" example:"пицца"`
	PromoCode     string               `json:"promo_code,omitempty" example:"WRD-1A2B3C4D"`
	PromoExpires  string               `json:"promo_expires_at,omitempty" example:"2026-06-03T00:00:00Z"`
	PromoDiscount int64                `json:"promo_discount_amount,omitempty" example:"100000000"`
}

func mapPBStatus(s pb.GameStatus) string {
	switch s {
	case pb.GameStatus_GAME_STATUS_PLAYING:
		return "PLAYING"
	case pb.GameStatus_GAME_STATUS_WON:
		return "WON"
	case pb.GameStatus_GAME_STATUS_LOST:
		return "LOST"
	default:
		return "UNSPECIFIED"
	}
}

func mapPBLetters(letters []pb.LetterState) []string {
	res := make([]string, len(letters))
	for i, l := range letters {
		switch l {
		case pb.LetterState_LETTER_STATE_CORRECT:
			res[i] = "CORRECT"
		case pb.LetterState_LETTER_STATE_PRESENT:
			res[i] = "PRESENT"
		case pb.LetterState_LETTER_STATE_ABSENT:
			res[i] = "ABSENT"
		default:
			res[i] = "UNSPECIFIED"
		}
	}
	return res
}

type GameHandler struct {
	gameClient  gameclient.GameClient
	userClient  userclient.UserClient
	promoClient pbOrder.PromoServiceClient
	logger      logger.Logger
	sanitizer   *bluemonday.Policy
}

func NewGameHandler(gc gameclient.GameClient, uc userclient.UserClient, pc pbOrder.PromoServiceClient, l logger.Logger) *GameHandler {
	return &GameHandler{
		gameClient:  gc,
		userClient:  uc,
		promoClient: pc,
		logger:      l,
		sanitizer:   bluemonday.UGCPolicy(),
	}
}

// GetDailyWordleState godoc
// @Summary 		Получение статуса игры "5 букв" за сегодня
// @Description		Возвращает текущий статус игры (PLAYING, WON, LOST), лимиты и историю попыток.
// @Tags			game
// @Produce			json
// @Success			200		{object}  DailyStateResponse		"Успешное получение статуса"
// @Failure			401		{object}  response.ErrorResponse	"Пользователь не авторизован"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/game/wordle [get]
func (h *GameHandler) GetDailyWordleState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	state, err := h.gameClient.GetDailyWordleState(ctx, userID)
	if err != nil {
		l.Error("failed to get daily wordle state", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respDTO := DailyStateResponse{
		WordLength:    state.WordLength,
		MaxAttempts:   state.MaxAttempts,
		Status:        mapPBStatus(state.Status),
		Guesses:       make([]WordleGuessResultDTO, 0, len(state.Guesses)),
		CurrentStreak: state.CurrentStreak,
		TargetWord:    state.TargetWord,
	}

	for _, g := range state.Guesses {
		respDTO.Guesses = append(respDTO.Guesses, WordleGuessResultDTO{
			Word:    g.Word,
			Letters: mapPBLetters(g.Letters),
		})
	}

	response.JSON(w, http.StatusOK, respDTO)
}

// MakeWordleGuess godoc
// @Summary 		Сделать попытку в игре "5 букв"
// @Description		Отправляет слово на проверку. Требует передачи заголовка Idempotency-Key.
// @Tags			game
// @Accept			json
// @Produce			json
// @Param			Idempotency-Key header string 				true	"Ключ идемпотентности"
// @Param			input	body	  MakeWordleGuessRequest	true	"Слово-попытка (строго 5 букв)"
// @Success			200		{object}  MakeGuessResponse			"Результат обработки попытки"
// @Failure			400		{object}  response.ErrorResponse	"Неверный формат запроса, нет заголовка или слова нет в словаре"
// @Failure			401		{object}  response.ErrorResponse	"Пользователь не авторизован"
// @Failure			403		{object}  response.ErrorResponse	"Игра уже завершена на сегодня или лимит исчерпан"
// @Failure			409		{object}  response.ErrorResponse	"Конфликт идемпотентности (запрос уже обработан)"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/game/wordle/guess [post]
func (h *GameHandler) MakeWordleGuess(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var reqDTO MakeWordleGuessRequest
	if err := request.JSON(r, &reqDTO); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	reqDTO.Sanitize(h.sanitizer)

	res, err := h.gameClient.MakeWordleGuess(ctx, userID, reqDTO.Guess, idemKey)
	if err != nil {
		switch {
		case errors.Is(err, gameclient.ErrInvalidWordLength):
			response.Error(w, http.StatusBadRequest, "word length must be exactly 5 letters")
		case errors.Is(err, gameclient.ErrWordNotInDictionary):
			response.Error(w, http.StatusBadRequest, "word is not in the dictionary")
		case errors.Is(err, gameclient.ErrGameAlreadyFinished):
			response.Error(w, http.StatusForbidden, "game is already finished for today")
		case errors.Is(err, gameclient.ErrIdempotencyConflict):
			response.Error(w, http.StatusConflict, "request with this idempotency key already processed")
		default:
			l.Error("failed to make wordle guess", err)
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	respDTO := MakeGuessResponse{
		Status:        mapPBStatus(res.Status),
		CurrentStreak: res.CurrentStreak,
		TargetWord:    res.TargetWord,
	}

	if res.GuessResult != nil {
		respDTO.GuessResult = WordleGuessResultDTO{
			Word:    res.GuessResult.Word,
			Letters: mapPBLetters(res.GuessResult.Letters),
		}
	}

	// За победу выпускаем единичный промокод (как и Колесо). При проигрыше —
	// никакого подарка. В обоих случаях триггерим ачивки.
	if res.Status == pb.GameStatus_GAME_STATUS_WON {
		discount := WordleWinPromoDiscount
		promo, perr := h.promoClient.CreateAndBindWheelPromo(ctx, &pbOrder.CreateAndBindWheelPromoRequest{
			UserId:         userID,
			DiscountAmount: &discount,
			Title:          "Скидка 100 ₽ за «5 букв»",
		})
		if perr != nil {
			l.Error("failed to issue wordle win promo", perr)
		} else {
			respDTO.PromoCode = promo.Code
			if promo.ExpiresAt != nil {
				respDTO.PromoExpires = *promo.ExpiresAt
			}
			respDTO.PromoDiscount = discount
		}
	}

	isWin := res.Status == pb.GameStatus_GAME_STATUS_WON
	if res.Status == pb.GameStatus_GAME_STATUS_WON || res.Status == pb.GameStatus_GAME_STATUS_LOST {
		if hookErr := h.userClient.OnWordleResult(ctx, userID, isWin, res.TotalWins, res.CurrentStreak); hookErr != nil {
			l.Warn("OnWordleResult hook failed", logger.Err(hookErr))
		}
	}

	response.JSON(w, http.StatusOK, respDTO)
}
