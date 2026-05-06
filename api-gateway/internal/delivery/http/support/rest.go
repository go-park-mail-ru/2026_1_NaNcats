package support

//go:generate easyjson $GOFILE

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/supportclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
)

//easyjson:json
type CreateTicketRequest struct {
	ContactEmail string `json:"contact_email"`
	CategoryID   int64  `json:"category_id"`
	FirstMessage string `json:"first_message"`
	ClientMeta   string `json:"client_meta"`
}

func (r *CreateTicketRequest) Sanitize(p *bluemonday.Policy) {
	r.FirstMessage = p.Sanitize(r.FirstMessage)
}

//easyjson:json
type RateTicketRequest struct {
	Rating int `json:"rating"`
}

//easyjson:json
type ChangeStatusRequest struct {
	Status string `json:"status"`
}

func (r *ChangeStatusRequest) Sanitize(p *bluemonday.Policy) {
	r.Status = p.Sanitize(r.Status)
}

//easyjson:json
type ReassignTicketRequest struct {
	AgentID int64 `json:"agent_id"`
	Line    int   `json:"line"`
}

//easyjson:json
type SetAgentStatusRequest struct {
	Status string `json:"status"`
}

func (r *SetAgentStatusRequest) Sanitize(p *bluemonday.Policy) {
	r.Status = p.Sanitize(r.Status)
}

//easyjson:json
type TicketDTO struct {
	ID               int64  `json:"id"`
	PublicID         string `json:"public_id"`
	CategoryID       int64  `json:"category_id"`
	CurrentStatus    string `json:"current_status"`
	SupportLine      int    `json:"support_line"`
	AssigneeID       int64  `json:"assignee_id,omitempty"`
	ResolutionRating int    `json:"resolution_rating,omitempty"`
	CreatedAt        string `json:"created_at"`
}

//easyjson:json
type EventDTO struct {
	ID         int64           `json:"id"`
	TicketID   int64           `json:"ticket_id"`
	AuthorID   int64           `json:"author_id,omitempty"`
	AuthorRole string          `json:"author_role"`
	EventType  string          `json:"event_type"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  string          `json:"created_at"`
}

//easyjson:json
type CategoryDTO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DefaultLine int    `json:"default_line"`
}

//easyjson:json
type TemplateDTO struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// Вспомогательные структуры для Swagger
//
//easyjson:json
type CreateTicketResponse struct {
	TicketID string `json:"ticket_id"`
}

//easyjson:json
type SuccessResponse struct {
	Success bool `json:"success"`
}

// Вспомогательные алиасы списков для удобной генерации MarshalJSON
//
//easyjson:json
type TicketListDTO []TicketDTO

//easyjson:json
type EventListDTO []EventDTO

//easyjson:json
type CategoryListDTO []CategoryDTO

//easyjson:json
type TemplateListDTO []TemplateDTO
type SupportHandler struct {
	supportClient supportclient.SupportClient
	redisHub      *RedisHub
	logger        logger.Logger
}

func NewSupportHandler(sc supportclient.SupportClient, rh *RedisHub, l logger.Logger) *SupportHandler {
	return &SupportHandler{
		supportClient: sc,
		redisHub:      rh,
		logger:        l,
	}
}

// CreateTicket godoc
// @Summary 		Создание тикета
// @Description		Создает новое обращение в службу поддержки (для авторизованных и неавторизованных пользователей)
// @Tags			support
// @Accept			json
// @Produce			json
// @Param			Idempotency-Key header string true "Ключ идемпотентности"
// @Param			input	body	  CreateTicketRequest	true	"Данные тикета"
// @Success			200		{object}  CreateTicketResponse			"Успешное создание (возвращает ticket_id)"
// @Failure			400		{object}  response.ErrorResponse	"Ошибка в формате запроса или отсутствие заголовка"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/support/tickets [post]
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
		ClientMeta:   json.RawMessage([]byte(req.ClientMeta)),
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

// GetMyTickets godoc
// @Summary 		Получение списка тикетов пользователя
// @Description		Возвращает все обращения текущего пользователя (идентифицируется по токену или куке guest_id)
// @Tags			support
// @Produce			json
// @Success			200		{array}   TicketDTO	"Список тикетов пользователя"
// @Failure			500		{object}  response.ErrorResponse		"Внутренняя ошибка сервера"
// @Router			/support/tickets [get]
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

	resp := make(TicketListDTO, 0, len(tickets))
	for _, t := range tickets {
		resp = append(resp, mapTicketToDTO(t))
	}

	response.JSON(w, http.StatusOK, resp)
}

// GetTicketEvents godoc
// @Summary 		Получение событий тикета
// @Description		Возвращает историю переписки и событий по конкретному тикету
// @Tags			support
// @Produce			json
// @Param			id		path	  string	true	"Public ID тикета"
// @Success			200		{array}   EventDTO	"Список событий"
// @Failure			403		{object}  response.ErrorResponse		"Доступ запрещен"
// @Failure			404		{object}  response.ErrorResponse		"Тикет не найден"
// @Failure			500		{object}  response.ErrorResponse		"Внутренняя ошибка сервера"
// @Router			/support/tickets/{id}/events [get]
func (h *SupportHandler) GetTicketEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ticketPublicID := r.PathValue("id")

	var clientID *int64
	var guestID *string

	uID, isAuth := middleware.GetUserID(ctx)
	if isAuth {
		clientID = &uID
	} else {
		gID := h.getOrSetGuestID(w, r)
		guestID = &gID
	}

	events, err := h.supportClient.GetTicketEvents(ctx, ticketPublicID, clientID, guestID)
	if err != nil {
		if errors.Is(err, supportclient.ErrTicketNotFound) {
			response.Error(w, http.StatusNotFound, "Ticket not found")
			return
		}
		if errors.Is(err, supportclient.ErrUnauthorized) {
			response.Error(w, http.StatusForbidden, "Access denied")
			return
		}
		h.logger.Error("failed to get ticket events", err)
		response.Error(w, http.StatusInternalServerError, "Failed to fetch events")
		return
	}

	resp := make(EventListDTO, 0, len(events))
	for _, e := range events {
		resp = append(resp, mapEventToDTO(e))
	}

	response.JSON(w, http.StatusOK, resp)
}

// RateTicket godoc
// @Summary 		Оценка решения тикета
// @Description		Позволяет пользователю поставить оценку решенному тикету
// @Tags			support
// @Accept			json
// @Produce			json
// @Param			id				path	  string	true	"Public ID тикета"
// @Param			Idempotency-Key header string true "Ключ идемпотентности"
// @Param			input			body	  RateTicketRequest	true	"Оценка тикета"
// @Success			200				{object}  SuccessResponse			"Успешная оценка"
// @Failure			400				{object}  response.ErrorResponse	"Ошибка в формате запроса"
// @Failure			500				{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/support/tickets/{id}/rate [post]
func (h *SupportHandler) RateTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ticketPublicID := r.PathValue("id")

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var req RateTicketRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var clientID *int64
	uID, isAuth := middleware.GetUserID(ctx)
	if isAuth {
		clientID = &uID
	}

	err := h.supportClient.RateTicket(ctx, ticketPublicID, req.Rating, clientID, idemKey)
	if err != nil {
		h.logger.Error("failed to rate ticket", err)
		response.Error(w, http.StatusInternalServerError, "Failed to rate ticket")
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GetAssignedTickets godoc
// @Summary 		Получение назначенных тикетов (Для агентов)
// @Description		Возвращает список тикетов, назначенных на текущего агента поддержки
// @Tags			agent
// @Produce			json
// @Success			200		{array}   TicketDTO	"Список назначенных тикетов"
// @Failure			401		{object}  response.ErrorResponse		"Неавторизован"
// @Failure			500		{object}  response.ErrorResponse		"Внутренняя ошибка сервера"
// @Router			/agent/tickets [get]
func (h *SupportHandler) GetAssignedTickets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tickets, err := h.supportClient.GetAssignedTickets(ctx, agentID)
	if err != nil {
		h.logger.Error("failed to get assigned tickets", err)
		response.Error(w, http.StatusInternalServerError, "Failed to fetch tickets")
		return
	}

	resp := make(TicketListDTO, 0, len(tickets))
	for _, t := range tickets {
		resp = append(resp, mapTicketToDTO(t))
	}

	response.JSON(w, http.StatusOK, resp)
}

// ChangeTicketStatus godoc
// @Summary 		Изменение статуса тикета (Для агентов)
// @Description		Меняет статус указанного тикета
// @Tags			agent
// @Accept			json
// @Produce			json
// @Param			id				path	  string	true	"Public ID тикета"
// @Param			Idempotency-Key header string true "Ключ идемпотентности"
// @Param			input			body	  ChangeStatusRequest	true	"Новый статус"
// @Success			200				{object}  SuccessResponse			"Статус успешно изменен"
// @Failure			400				{object}  response.ErrorResponse	"Ошибка запроса"
// @Failure			401				{object}  response.ErrorResponse	"Неавторизован"
// @Failure			500				{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/agent/tickets/{id}/status [patch]
func (h *SupportHandler) ChangeTicketStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ticketPublicID := r.PathValue("id")

	agentID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var req ChangeStatusRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.supportClient.ChangeTicketStatus(ctx, ticketPublicID, req.Status, agentID, idemKey)
	if err != nil {
		h.logger.Error("failed to change ticket status", err)
		response.Error(w, http.StatusInternalServerError, "Failed to update status")
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

// ReassignTicket godoc
// @Summary 		Переназначение тикета (Для агентов)
// @Description		Передает тикет другому агенту или на другую линию поддержки
// @Tags			agent
// @Accept			json
// @Produce			json
// @Param			id				path	  string	true	"Public ID тикета"
// @Param			Idempotency-Key header string true "Ключ идемпотентности"
// @Param			input			body	  ReassignTicketRequest	true	"Данные для переназначения"
// @Success			200				{object}  SuccessResponse			"Тикет успешно переназначен"
// @Failure			400				{object}  response.ErrorResponse	"Ошибка запроса"
// @Failure			401				{object}  response.ErrorResponse	"Неавторизован"
// @Failure			500				{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/agent/tickets/{id}/reassign [post]
func (h *SupportHandler) ReassignTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ticketPublicID := r.PathValue("id")

	authorID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var req ReassignTicketRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.supportClient.ReassignTicket(ctx, ticketPublicID, req.AgentID, req.Line, authorID, idemKey)
	if err != nil {
		h.logger.Error("failed to reassign ticket", err)
		response.Error(w, http.StatusInternalServerError, "Failed to reassign ticket")
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

// SetAgentStatus godoc
// @Summary 		Установка рабочего статуса агента
// @Description		Позволяет агенту изменить свой статус (например: активен, отошел)
// @Tags			agent
// @Accept			json
// @Produce			json
// @Param			input	body	  SetAgentStatusRequest	true	"Новый статус агента"
// @Success			200		{object}  SuccessResponse			"Статус успешно обновлен"
// @Failure			400		{object}  response.ErrorResponse	"Ошибка запроса"
// @Failure			401		{object}  response.ErrorResponse	"Неавторизован"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/agent/status [patch]
func (h *SupportHandler) SetAgentStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req SetAgentStatusRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.supportClient.SetAgentStatus(ctx, agentID, req.Status)
	if err != nil {
		h.logger.Error("failed to set agent status", err)
		response.Error(w, http.StatusInternalServerError, "Failed to update agent status")
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GetCategories godoc
// @Summary 		Получение категорий обращений
// @Description		Возвращает список доступных категорий для создания тикетов
// @Tags			support
// @Produce			json
// @Success			200		{array}   CategoryDTO	"Список категорий"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/support/categories [get]
func (h *SupportHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	categories, err := h.supportClient.GetCategories(ctx)
	if err != nil {
		h.logger.Error("failed to get categories", err)
		response.Error(w, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	resp := make(CategoryListDTO, 0, len(categories))
	for _, c := range categories {
		resp = append(resp, CategoryDTO{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			DefaultLine: c.DefaultLine,
		})
	}
	response.JSON(w, http.StatusOK, resp)
}

// GetTemplates godoc
// @Summary 		Получение шаблонов ответов (Для агентов)
// @Description		Возвращает список готовых шаблонов(отбивок) для использования агентами в чате
// @Tags			agent
// @Produce			json
// @Success			200		{array}   TemplateDTO	"Список шаблонов"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/agent/templates [get]
func (h *SupportHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	templates, err := h.supportClient.GetTemplates(ctx)
	if err != nil {
		h.logger.Error("failed to get templates", err)
		response.Error(w, http.StatusInternalServerError, "Failed to fetch templates")
		return
	}

	resp := make(TemplateListDTO, 0, len(templates))
	for _, t := range templates {
		resp = append(resp, TemplateDTO{
			ID:      t.ID,
			Name:    t.Name,
			Content: t.Content,
		})
	}
	response.JSON(w, http.StatusOK, resp)
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

// GetSupportStats godoc
// @Summary      Получение статистики техподдержки
// @Description  Возвращает агрегированные данные по тикетам: общее количество, распределение по статусам и категориям, средний рейтинг и среднее время решения. Доступно только для ролей admin и support.
// @Tags         admin, support
// @Produce      json
// @Success      200  {object}  supportclient.SupportStats "Статистика успешно получена"
// @Failure      401  {object}  response.ErrorResponse     "Пользователь не авторизован"
// @Failure      403  {object}  response.ErrorResponse     "Недостаточно прав (требуется роль admin или support)"
// @Failure      500  {object}  response.ErrorResponse     "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /admin/support/stats [get]
func (h *SupportHandler) GetSupportStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	role, ok := ctx.Value(middleware.RoleKey).(string)
	if !ok || (role != "admin" && role != "support") {
		response.Error(w, http.StatusForbidden, "Access denied: admin or support role required")
		return
	}

	stats, err := h.supportClient.GetStats(ctx)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, stats)
}

func mapTicketToDTO(t supportclient.Ticket) TicketDTO {
	return TicketDTO{
		ID:               t.ID,
		PublicID:         t.PublicID,
		CategoryID:       t.CategoryID,
		CurrentStatus:    t.CurrentStatus,
		SupportLine:      t.SupportLine,
		AssigneeID:       t.AssigneeID,
		ResolutionRating: t.ResolutionRating,
		CreatedAt:        t.CreatedAt.Format(time.RFC3339),
	}
}

func mapEventToDTO(e supportclient.Event) EventDTO {
	return EventDTO{
		ID:         e.ID,
		TicketID:   e.TicketID,
		AuthorID:   e.AuthorID,
		AuthorRole: e.AuthorRole,
		EventType:  e.EventType,
		Payload:    e.Payload,
		CreatedAt:  e.CreatedAt.Format(time.RFC3339),
	}
}
