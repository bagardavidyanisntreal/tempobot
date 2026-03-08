package model

import "time"

type Event struct {
	ID int64

	Title       string
	Description string

	StartAt              time.Time
	RegistrationDeadline time.Time

	CreatedBy int64

	ChatID    int64
	MessageID int
}
