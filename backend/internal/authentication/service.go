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

	ErrUserNotFound = _package.NewAppError(
		"User not found",
		_package.CodeNotFound,
		"User does not exist",
	)

	ErrCurrentPasswordNotMatch = _package.NewAppError(
		"Current password does not match",
		_package.CodeValidation,
		"The current password you entered is incorrect",
	)
)

type Service interface {
	Login(ctx context.Context, username string, password string) (*models.User, error)
	ChangePassword(ctx context.Context, userId string, currentPassword string, newPassword string) error
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

func (s *service) ChangePassword(ctx context.Context, username string, currentPassword string, newPassword string) error {

	user, err := s.repository.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("get user by username: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return ErrCurrentPasswordNotMatch
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}
	hashPassword := string(hash)

	if err := s.repository.UpdatePassword(ctx, username, hashPassword); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}
