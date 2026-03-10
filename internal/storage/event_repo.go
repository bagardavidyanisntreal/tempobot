package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bagardavidyanisntreal/tempobot/internal/model"
)

type EventRepo struct {
	db *sql.DB
}

func NewEventRepo(db *sql.DB) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) Get(ctx context.Context, id int64) (*model.Event, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id,title,description,chat_id,message_id
		FROM events
		WHERE id=$1
	`, id)

	var e model.Event

	err := row.Scan(
		&e.ID,
		&e.Title,
		&e.Description,
		&e.ChatID,
		&e.MessageID,
	)
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}

	return &e, nil
}

func (r *EventRepo) Create(
	ctx context.Context,
	title string,
	description string,
	chatID int64,
) (int64, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO events (title, description, chat_id)
		VALUES ($1,$2,$3)
		RETURNING id
	`, title, description, chatID)

	var id int64

	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("create event: %w", err)
	}

	return id, nil
}

func (r *EventRepo) AttachMessage(
	ctx context.Context,
	eventID int64,
	messageID int64,
) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE events
		SET message_id = $1
		WHERE id = $2
	`, messageID, eventID)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}

	return nil
}
