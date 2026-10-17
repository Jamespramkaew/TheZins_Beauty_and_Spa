package authentication

import (
	"context"
	_package "backend/package"
	"backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)
var (
	ErrInvalidCredentials = _package.NewAppError(
		"Invalid credentials",
		"UNAUTHORIZED",
		"Username or password is incorrect",
	)
	ErrUserNotFound = _package.NewAppError(
		"User not found",
		"NOT_FOUND",
		"Username is invalid",
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
		return nil, ErrUserNotFound
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

