package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kurushqosimi/x5-intern-hiring/internal/models"
	"github.com/kurushqosimi/x5-intern-hiring/internal/repositories"
)

func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(ids))
	for _, s := range ids {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func (s *Service) Invite(ctx context.Context, req models.BulkEmailActionRequest) (models.BulkEmailActionResponse, error) {
	if req.TemplateCode == "" {
		req.TemplateCode = "intern_invite_v1"
	}
	ids, err := parseUUIDs(req.ApplicationIDs)
	if err != nil {
		return models.BulkEmailActionResponse{}, err
	}
	return s.repo.QueueInviteEmails(ctx, ids, req.TemplateCode)
}

func (s *Service) Reject(ctx context.Context, req models.BulkEmailActionRequest) (models.BulkEmailActionResponse, error) {
	if req.TemplateCode == "" {
		req.TemplateCode = "intern_reject_v1"
	}
	ids, err := parseUUIDs(req.ApplicationIDs)
	if err != nil {
		return models.BulkEmailActionResponse{}, err
	}
	return s.repo.QueueRejectEmails(ctx, ids, req.TemplateCode, req.StatusReason)
}

func IsTemplateNotFound(err error) bool {
	return errors.Is(err, repositories.ErrTemplateNotFound)
}
