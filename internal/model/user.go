package model

type User struct {
	ID int64

	TelegramUserID int64
	Username       string
	FirstName      string

	Role string
}
