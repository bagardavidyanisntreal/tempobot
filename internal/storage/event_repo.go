package storage

import (
	"database/sql"

	"github.com/bagardavidyanisntreal/tempobot/internal/model"
)

type EventRepo struct {
	db *sql.DB
}

func NewEventRepo(db *sql.DB) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) Get(id int64) (*model.Event, error) {
	row := r.db.QueryRow(`
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
		return nil, err
	}

	return &e, nil
}

func (r *EventRepo) Create(
	title string,
	description string,
	chatID int64,
) (int64, error) {

	row := r.db.QueryRow(`
		INSERT INTO events (title, description, chat_id)
		VALUES ($1,$2,$3)
		RETURNING id
	`, title, description, chatID)

	var id int64

	err := row.Scan(&id)

	return id, err
}
