package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/kurushqosimi/x5-intern-hiring/internal/models"
)

func (s *Service) QueueToCRM(ctx context.Context, req models.BulkCRMActionRequest) (models.BulkCRMActionResponse, error) {
	ids := make([]uuid.UUID, 0, len(req.ApplicationIDs))
	for _, x := range req.ApplicationIDs {
		id, err := uuid.Parse(x)
		if err != nil {
			return models.BulkCRMActionResponse{}, err
		}
		ids = append(ids, id)
	}
	return s.repo.QueueCRM(ctx, ids)
}
