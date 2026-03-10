package service

import (
	"context"
	"fmt"

	"github.com/bagardavidyanisntreal/tempobot/internal/storage"
)

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

func (s *EventService) RegisterParticipant(ctx context.Context, eventID, userID int64, status string) error {
	if err := s.partRepo.Upsert(ctx, eventID, userID, status); err != nil {
		return fmt.Errorf("upsert participant: %w", err)
	}

	return nil
}

func (s *EventService) GetEvent(ctx context.Context, eventID int64) (any, error) {
	event, err := s.eventRepo.Get(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}

	return event, nil
}

func (s *EventService) GetStats(ctx context.Context, eventID int64) (map[string]int, error) {
	stats, err := s.partRepo.CountByStatus(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}

	return stats, nil
}

func (s *EventService) CreateEvent(ctx context.Context, title string, description string, chatID int64) (int64, error) {
	id, err := s.eventRepo.Create(ctx, title, description, chatID)
	if err != nil {
		return 0, fmt.Errorf("create event: %w", err)
	}

	return id, nil
}

func (s *EventService) AttachMessage(
	ctx context.Context,
	eventID int64,
	messageID int64,
) error {
	if err := s.eventRepo.AttachMessage(ctx, eventID, messageID); err != nil {
		return fmt.Errorf("attach message: %w", err)
	}

	return nil
}
