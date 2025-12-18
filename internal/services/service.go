package services

import "github.com/kurushqosimi/x5-intern-hiring/internal/repositories"

type Service struct {
	repo *repositories.Repository
}

func NewService(repo *repositories.Repository) *Service {
	return &Service{repo: repo}
}
