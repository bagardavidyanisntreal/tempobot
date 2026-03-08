package service

import (
	"database/sql"
	"errors"

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
	telegramID int64,
	username string,
	firstName string,
) (*model.User, error) {
	u, err := s.repo.FindByTelegramID(telegramID)
	if err == nil {
		_ = s.repo.UpdateProfile(u.ID, username, firstName)

		return u, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return s.repo.Create(telegramID, username, firstName)
	}

	return nil, err
}
