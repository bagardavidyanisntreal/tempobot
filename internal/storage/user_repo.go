package storage

import (
	"database/sql"

	"github.com/bagardavidyanisntreal/tempobot/internal/model"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByTelegramID(id int64) (*model.User, error) {
	row := r.db.QueryRow(`
		SELECT id, telegram_user_id, username, first_name, role
		FROM users
		WHERE telegram_user_id = $1
	`, id)

	var u model.User

	err := row.Scan(
		&u.ID,
		&u.TelegramUserID,
		&u.Username,
		&u.FirstName,
		&u.Role,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepo) Create(
	telegramID int64,
	username string,
	firstName string,
) (*model.User, error) {
	row := r.db.QueryRow(`
		INSERT INTO users (telegram_user_id, username, first_name)
		VALUES ($1,$2,$3)
		RETURNING id, telegram_user_id, username, first_name, role
	`, telegramID, username, firstName)

	var u model.User

	err := row.Scan(
		&u.ID,
		&u.TelegramUserID,
		&u.Username,
		&u.FirstName,
		&u.Role,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepo) UpdateProfile(
	id int64,
	username string,
	firstName string,
) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET username = $1,
			first_name = $2
		WHERE id = $3
	`, username, firstName, id)

	return err
}
