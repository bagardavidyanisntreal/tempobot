package service

import "github.com/bagardavidyanisntreal/tempobot/internal/storage"

type EventService struct {
	eventRepo *storage.EventRepo
	partRepo  *storage.ParticipantRepo
}

func NewEventService(
	e *storage.EventRepo,
	p *storage.ParticipantRepo,
) *EventService {
	return &EventService{
		eventRepo: e,
		partRepo:  p,
	}
}

func (s *EventService) RegisterParticipant(eventID, userID int64, status string) error {
	return s.partRepo.Upsert(eventID, userID, status)
}

func (s *EventService) GetEvent(eventID int64) (interface{}, error) {
	return s.eventRepo.Get(eventID)
}

func (s *EventService) GetStats(eventID int64) (map[string]int, error) {
	return s.partRepo.CountByStatus(eventID)
}

func (s *EventService) CreateEvent(
	title string,
	description string,
	chatID int64,
) (int64, error) {

	return s.eventRepo.Create(
		title,
		description,
		chatID,
	)
}
