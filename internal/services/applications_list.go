package services

import (
	"context"
	"github.com/kurushqosimi/x5-intern-hiring/internal/models"
)

func (s *Service) ListApplications(ctx context.Context, p models.ListApplicationsParams) (models.ListApplicationsResponse, error) {
	items, total, err := s.repo.ListApplications(ctx, p)
	if err != nil {
		return models.ListApplicationsResponse{}, err
	}
	return models.ListApplicationsResponse{
		Items:  items,
		Total:  total,
		Limit:  p.Limit,
		Offset: p.Offset,
	}, nil
}
