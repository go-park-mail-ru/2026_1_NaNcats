package support

//go:generate easyjson $GOFILE

import (
	"encoding/json"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/supportclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
	"github.com/google/uuid"
)

//easyjson:json
type CreateTicketRequest struct {
	ContactEmail string `json:"contact_email"`
	CategoryID   int64  `json:"category_id"`
	FirstMessage string `json:"first_message"`
	ClientMeta   string `json:"client_meta"`
}

type SupportHandler struct {
	supportClient supportclient.SupportClient
	logger        logger.Logger
}

func NewSupportHandler(sc supportclient.SupportClient, l logger.Logger) *SupportHandler {
	return &SupportHandler{
		supportClient: sc,
		logger:        l,
	}
}

func (h *SupportHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var req CreateTicketRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := supportclient.CreateTicketInput{
		ContactEmail: req.ContactEmail,
		CategoryID:   req.CategoryID,
		FirstMessage: req.FirstMessage,
		ClientMeta:   json.RawMessage(req.ClientMeta),
	}

	userID, isAuth := middleware.GetUserID(ctx)
	if isAuth {
		input.ClientID = &userID
	} else {
		guestID := h.getOrSetGuestID(w, r)
		input.GuestID = &guestID
	}

	ticketPublicID, err := h.supportClient.CreateTicket(ctx, input, idemKey)
	if err != nil {
		l.Error("ticket creation failed", err)
		response.Error(w, http.StatusInternalServerError, "Failed to create ticket")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"ticket_id": ticketPublicID})
}

func (h *SupportHandler) GetMyTickets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var clientID *int64
	var guestID *string

	uID, isAuth := middleware.GetUserID(ctx)
	if isAuth {
		clientID = &uID
	} else {
		gID := h.getOrSetGuestID(w, r)
		guestID = &gID
	}

	tickets, err := h.supportClient.GetUserTickets(ctx, clientID, guestID)
	if err != nil {
		h.logger.Error("failed to get tickets", err)
		response.Error(w, http.StatusInternalServerError, "Failed to fetch tickets")
		return
	}

	response.JSON(w, http.StatusOK, tickets)
}

func (h *SupportHandler) getOrSetGuestID(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("guest_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	guestID := uuid.New().String()
	http.SetCookie(w, &http.Cookie{
		Name:     "guest_id",
		Value:    guestID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   86400 * 30,
	})
	return guestID
}
