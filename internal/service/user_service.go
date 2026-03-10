package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bagardavidyanisntreal/tempobot/internal/model"
	"github.com/bagardavidyanisntreal/tempobot/internal/storage"
)

type UserService struct {
	repo *storage.UserRepo
}

func NewUserService(r *storage.UserRepo) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) EnsureUser(
	ctx context.Context,
	telegramID int64,
	username string,
	firstName string,
) (*model.User, error) {
	u, err := s.repo.FindByTelegramID(ctx, telegramID)
	if err == nil {
		_ = s.repo.UpdateProfile(ctx, u.ID, username, firstName)

		return u, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		u, err = s.repo.Create(ctx, telegramID, username, firstName)
		if err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
	}

	return u, nil
}
