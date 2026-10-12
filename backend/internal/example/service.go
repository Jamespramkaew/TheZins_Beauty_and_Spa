package example

import (
	"context"
	"backend/database"
)

type Service interface {
	GetAllExampleMessage(ctx context.Context, limit, offset int) ([]database.Example, int64, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

func (s *service) GetAllExampleMessage(ctx context.Context, limit, offset int) ([]database.Example, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	message, totalCount, err := s.repository.GetAllExampleMessage(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return message, totalCount, nil
}