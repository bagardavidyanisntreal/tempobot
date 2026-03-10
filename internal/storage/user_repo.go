package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bagardavidyanisntreal/tempobot/internal/model"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByTelegramID(ctx context.Context, id int64) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `
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
		return nil, fmt.Errorf("FindByTelegramID: %w", err)
	}

	return &u, nil
}

func (r *UserRepo) Create(ctx context.Context, telegramID int64, username string, firstName string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `
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
		return nil, fmt.Errorf("Create: %w", err)
	}

	return &u, nil
}

func (r *UserRepo) UpdateProfile(ctx context.Context, id int64, username string, firstName string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET username = $1,
			first_name = $2
		WHERE id = $3
	`, username, firstName, id)
	if err != nil {
		return fmt.Errorf("UpdateProfile: %w", err)
	}

	return nil
}
