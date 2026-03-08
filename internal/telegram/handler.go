package telegram

import (
	"encoding/json"
	"log"
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
		http.Error(w, err.Error(), 400)
		return
	}

	if upd.CallbackQuery != nil {
		h.handleCallback(upd.CallbackQuery)
	}

	if upd.Message != nil {
		h.handleMessage(upd.Message)
	}

	w.WriteHeader(200)
}

func (h *Handler) handleCallback(cb *CallbackQuery) {
	user := cb.From

	u, err := h.userService.EnsureUser(
		user.ID,
		user.Username,
		user.FirstName,
	)

	if err != nil {
		log.Println(err)
		return
	}

	parts := strings.Split(cb.Data, ":")

	if len(parts) != 3 {
		return
	}

	eventID, _ := strconv.ParseInt(parts[1], 10, 64)

	status := parts[2]

	err = h.eventService.RegisterParticipant(eventID, u.ID, status)
	if err != nil {
		log.Println(err)
		return
	}

	eRaw, err := h.eventService.GetEvent(eventID)
	if err != nil {
		log.Println(err)
		return
	}

	event := eRaw.(*model.Event)

	stats, err := h.eventService.GetStats(eventID)
	if err != nil {
		log.Println(err)
		return
	}

	text := BuildEventText(event, stats)

	err = h.tg.EditMessage(
		event.ChatID,
		event.MessageID,
		text,
		eventID,
		stats,
	)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) handleMessage(msg *Message) {

	if !strings.HasPrefix(msg.Text, "/create") {
		return
	}

	parts := strings.Split(msg.Text, "|")

	if len(parts) < 3 {
		h.tg.SendMessage(msg.Chat.ID, "формат: /create | title | description")
		return
	}

	title := strings.TrimSpace(parts[1])
	description := strings.TrimSpace(parts[2])

	eventID, err := h.eventService.CreateEvent(
		title,
		description,
		msg.Chat.ID,
	)

	if err != nil {
		return
	}

	h.publishEvent(eventID)
}

func (h *Handler) publishEvent(eventID int64) error {

	eventRaw, err := h.eventService.GetEvent(eventID)
	if err != nil {
		return err
	}

	event := eventRaw.(*model.Event)

	stats := map[string]int{
		"going": 0,
		"maybe": 0,
		"no":    0,
	}

	text := BuildEventText(event, stats)

	msg, err := h.tg.SendEventMessage(
		event.ChatID,
		text,
		eventID,
		stats,
	)

	if err != nil {
		return err
	}

	return h.eventService.AttachMessage(eventID, msg.MessageID)
}
