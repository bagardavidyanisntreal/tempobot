package model

type Participant struct {
	ID int64

	EventID int64
	UserID  int64

	Status string
}
