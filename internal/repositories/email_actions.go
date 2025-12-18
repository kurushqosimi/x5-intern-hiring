package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kurushqosimi/x5-intern-hiring/internal/models"
)

var ErrTemplateNotFound = errors.New("template not found")

func (repo *Repository) getTemplateIDByCode(ctx context.Context, tx pgx.Tx, code string) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT template_id
		FROM message_templates
		WHERE code=$1 AND is_active=true
		LIMIT 1
	`, code).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrTemplateNotFound
		}
		return uuid.Nil, err
	}
	return id, nil
}

type appRow struct {
	AppID     uuid.UUID
	CandID    uuid.UUID
	Status    string
	FirstName string
	Email     string
}

func (repo *Repository) QueueInviteEmails(ctx context.Context, appIDs []uuid.UUID, templateCode string) (models.BulkEmailActionResponse, error) {
	return repo.queueEmails(ctx, appIDs, templateCode, models.AppInviteQueued, nil, []string{models.AppInviteQueued, models.AppInvited})
}

func (repo *Repository) QueueRejectEmails(ctx context.Context, appIDs []uuid.UUID, templateCode string, reason string) (models.BulkEmailActionResponse, error) {
	var r *string
	if reason != "" {
		r = &reason
	}
	return repo.queueEmails(ctx, appIDs, templateCode, models.AppRejectQueued, r, []string{models.AppRejectQueued, models.AppRejected})
}

func (repo *Repository) queueEmails(
	ctx context.Context,
	appIDs []uuid.UUID,
	templateCode string,
	newStatus string,
	statusReason *string,
	skipStatuses []string,
) (models.BulkEmailActionResponse, error) {

	if len(appIDs) == 0 {
		return models.BulkEmailActionResponse{}, nil
	}

	tx, err := repo.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.BulkEmailActionResponse{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	tplID, err := repo.getTemplateIDByCode(ctx, tx, templateCode)
	if err != nil {
		return models.BulkEmailActionResponse{}, err
	}

	// Получаем: статус + first_name + email
	rows, err := tx.Query(ctx, `
		SELECT
			a.application_id,
			a.candidate_id,
			a.status,
			c.first_name,
			COALESCE((
				SELECT cc.value
				FROM candidate_contacts cc
				WHERE cc.candidate_id = a.candidate_id AND cc.type = 'email'
				ORDER BY cc.is_primary DESC, cc.created_at DESC
				LIMIT 1
			), '') AS email
		FROM applications a
		JOIN candidates c ON c.candidate_id = a.candidate_id
		WHERE a.application_id = ANY($1::uuid[])
	`, appIDs)
	if err != nil {
		return models.BulkEmailActionResponse{}, err
	}

	var appRows []appRow
	for rows.Next() {
		var r appRow
		if err := rows.Scan(&r.AppID, &r.CandID, &r.Status, &r.FirstName, &r.Email); err != nil {
			rows.Close()
			return models.BulkEmailActionResponse{}, err
		}
		appRows = append(appRows, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return models.BulkEmailActionResponse{}, err
	}
	rows.Close()

	skipSet := map[string]struct{}{}
	for _, st := range skipStatuses {
		skipSet[st] = struct{}{}
	}

	res := models.BulkEmailActionResponse{}

	for _, r := range appRows {
		// уже в очереди/отправлено — пропуск
		if _, ok := skipSet[r.Status]; ok {
			res.Skipped++
			continue
		}

		// нет email — пропуск + ошибка
		if r.Email == "" {
			res.Skipped++
			res.Errors = append(res.Errors, models.ActionItemError{
				ApplicationID: r.AppID.String(),
				Error:         "у кандидата отсутствует email",
			})
			continue
		}

		vars := map[string]any{
			"first_name":     r.FirstName,
			"application_id": r.AppID.String(),
		}
		b, _ := json.Marshal(vars)
		renderVarsJSON := string(b)

		// кладем в outbox (render_vars jsonb)
		_, exErr := tx.Exec(ctx, `
			INSERT INTO email_outbox(email_id, application_id, to_email, template_id, render_vars, status)
			VALUES($1,$2,$3,$4,$5::jsonb,'PENDING')
		`, uuid.New(), r.AppID, r.Email, tplID, renderVarsJSON)
		if exErr != nil {
			res.Skipped++
			res.Errors = append(res.Errors, models.ActionItemError{
				ApplicationID: r.AppID.String(),
				Error:         fmt.Sprintf("не удалось добавить в outbox: %v", exErr),
			})
			continue
		}

		// обновляем статус заявки
		if statusReason != nil {
			_, exErr = tx.Exec(ctx, `
				UPDATE applications
				SET status=$2, status_reason=$3, updated_at=now()
				WHERE application_id=$1
			`, r.AppID, newStatus, *statusReason)
		} else {
			_, exErr = tx.Exec(ctx, `
				UPDATE applications
				SET status=$2, updated_at=now()
				WHERE application_id=$1
			`, r.AppID, newStatus)
		}
		if exErr != nil {
			res.Skipped++
			res.Errors = append(res.Errors, models.ActionItemError{
				ApplicationID: r.AppID.String(),
				Error:         fmt.Sprintf("не удалось обновить статус: %v", exErr),
			})
			continue
		}

		res.Queued++
	}

	if rows.Err() != nil {
		return models.BulkEmailActionResponse{}, rows.Err()
	}

	err = tx.Commit(ctx)
	if err != nil {
		return models.BulkEmailActionResponse{}, err
	}
	return res, nil
}
