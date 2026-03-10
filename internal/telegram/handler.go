package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bagardavidyanisntreal/tempobot/internal/model"
	"github.com/bagardavidyanisntreal/tempobot/internal/service"
)

type Handler struct {
	userService  *service.UserService
	eventService *service.EventService
	tg           *Client
}

func NewHandler(
	us *service.UserService,
	es *service.EventService,
	tg *Client,
) *Handler {
	return &Handler{
		userService:  us,
		eventService: es,
		tg:           tg,
	}
}

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var upd Update

	err := json.NewDecoder(r.Body).Decode(&upd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if upd.CallbackQuery != nil {
		if err = h.handleCallback(r.Context(), upd.CallbackQuery); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	}

	if upd.Message != nil {
		if err = h.handleMessage(r.Context(), upd.Message); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleCallback(ctx context.Context, cb *CallbackQuery) error {
	user := cb.From

	u, err := h.userService.EnsureUser(ctx, user.ID, user.Username, user.FirstName)
	if err != nil {
		return fmt.Errorf("ensure user: %w", err)
	}

	parts := strings.Split(cb.Data, ":")

	if len(parts) != msgPartsCnt {
		return fmt.Errorf("invalid callback data")
	}

	eventID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid event id")
	}

	status := parts[2]

	err = h.eventService.RegisterParticipant(ctx, eventID, u.ID, status)
	if err != nil {
		return fmt.Errorf("register participant: %w", err)
	}

	eRaw, err := h.eventService.GetEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("get event: %w", err)
	}

	event, ok := eRaw.(*model.Event)
	if !ok {
		return fmt.Errorf("invalid event data")
	}

	stats, err := h.eventService.GetStats(ctx, eventID)
	if err != nil {
		return fmt.Errorf("get stats: %w", err)
	}

	text := BuildEventText(event, stats)

	err = h.tg.EditMessage(
		ctx,
		event.ChatID,
		event.MessageID,
		text,
		eventID,
		stats,
	)
	if err != nil {
		return fmt.Errorf("edit message: %w", err)
	}

	return nil
}

const msgPartsCnt = 3

func (h *Handler) handleMessage(ctx context.Context, msg *Message) error {
	if !strings.HasPrefix(msg.Text, "/create") {
		return nil
	}

	parts := strings.Split(msg.Text, "|")
	if len(parts) < msgPartsCnt {
		return h.tg.SendMessage(ctx, msg.Chat.ID, "формат: /create | title | description")
	}

	title := strings.TrimSpace(parts[1])
	description := strings.TrimSpace(parts[2])

	eventID, err := h.eventService.CreateEvent(ctx, title, description, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("create event: %w", err)
	}

	return h.publishEvent(ctx, eventID)
}

func (h *Handler) publishEvent(ctx context.Context, eventID int64) error {
	eventRaw, err := h.eventService.GetEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed get event: %w", err)
	}

	event, ok := eventRaw.(*model.Event)
	if !ok {
		return fmt.Errorf("raw event data is not an event")
	}

	stats := map[string]int{
		"going": 0,
		"maybe": 0,
		"no":    0,
	}

	text := BuildEventText(event, stats)

	msg, err := h.tg.SendEventMessage(
		ctx,
		event.ChatID,
		text,
		eventID,
		stats,
	)
	if err != nil {
		return err
	}

	if err = h.eventService.AttachMessage(ctx, eventID, msg.Result.MessageID); err != nil {
		return fmt.Errorf("attach msg to event: %w", err)
	}

	return nil
}
