package authentication

import (
	"context"
	"errors"
	"fmt"

	"backend/internal/models"
	_package "backend/package"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = _package.NewAppError(
		"Invalid credentials",
		_package.CodeUnauthorized,
		"Username or password is incorrect",
	)
)

type Service interface {
	Login(ctx context.Context, username string, password string) (*models.User, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) Login(ctx context.Context, username string, password string) (*models.User, error) {
	user, err := s.repository.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("get user by username: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
