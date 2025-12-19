package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kurushqosimi/x5-intern-hiring/internal/models"
)

type crmRow struct {
	AppID     uuid.UUID
	CandID    uuid.UUID
	Status    string
	AppliedAt any // timestamptz -> time.Time, но можно не тащить типы здесь (scan сам приведёт в time.Time)
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Telegram  string
	ResumeURL *string
	Priority1 *string
	Priority2 *string
	RawRow    []byte // jsonb в pgx обычно сканится в []byte
}

func (repo *Repository) QueueCRM(ctx context.Context, appIDs []uuid.UUID) (models.BulkCRMActionResponse, error) {
	if len(appIDs) == 0 {
		return models.BulkCRMActionResponse{}, nil
	}

	tx, err := repo.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.BulkCRMActionResponse{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// 1) читаем все строки (ВАЖНО: полностью вычитываем rows, потом закрываем, и только потом делаем INSERT/UPDATE)
	rows, err := tx.Query(ctx, `
		SELECT
			a.application_id,
			a.candidate_id,
			a.status,
			a.applied_at,
			c.first_name,
			c.last_name,
			COALESCE((
				SELECT cc.value FROM candidate_contacts cc
				WHERE cc.candidate_id=a.candidate_id AND cc.type='email'
				ORDER BY cc.is_primary DESC, cc.created_at DESC
				LIMIT 1
			), '') AS email,
			COALESCE((
				SELECT cc.value FROM candidate_contacts cc
				WHERE cc.candidate_id=a.candidate_id AND cc.type='phone'
				ORDER BY cc.is_primary DESC, cc.created_at DESC
				LIMIT 1
			), '') AS phone,
			COALESCE((
				SELECT cc.value FROM candidate_contacts cc
				WHERE cc.candidate_id=a.candidate_id AND cc.type='telegram'
				ORDER BY cc.is_primary DESC, cc.created_at DESC
				LIMIT 1
			), '') AS telegram,
			a.resume_url,
			a.priority1,
			a.priority2,
			a.raw_row
		FROM applications a
		JOIN candidates c ON c.candidate_id = a.candidate_id
		WHERE a.application_id = ANY($1::uuid[])
	`, appIDs)
	if err != nil {
		return models.BulkCRMActionResponse{}, err
	}

	var all []crmRow
	for rows.Next() {
		var r crmRow
		if scanErr := rows.Scan(
			&r.AppID,
			&r.CandID,
			&r.Status,
			&r.AppliedAt,
			&r.FirstName,
			&r.LastName,
			&r.Email,
			&r.Phone,
			&r.Telegram,
			&r.ResumeURL,
			&r.Priority1,
			&r.Priority2,
			&r.RawRow,
		); scanErr != nil {
			rows.Close()
			return models.BulkCRMActionResponse{}, scanErr
		}
		all = append(all, r)
	}
	if rows.Err() != nil {
		rows.Close()
		return models.BulkCRMActionResponse{}, rows.Err()
	}
	rows.Close()

	// 2) ставим в outbox + обновляем статус
	skipSet := map[string]struct{}{
		models.AppCRMQueued: {},
		models.AppCRMSynced: {},
	}

	res := models.BulkCRMActionResponse{}

	for _, r := range all {
		if _, ok := skipSet[r.Status]; ok {
			res.Skipped++
			continue
		}

		// payload — пока универсальный (под будущую CRM API)
		payload := map[string]any{
			"application_id": r.AppID.String(),
			"candidate": map[string]any{
				"candidate_id": r.CandID.String(),
				"first_name":   r.FirstName,
				"last_name":    r.LastName,
				"contacts": map[string]any{
					"email":    r.Email,
					"phone":    r.Phone,
					"telegram": r.Telegram,
				},
			},
			"resume_url": r.ResumeURL,
			"priority1":  r.Priority1,
			"priority2":  r.Priority2,
			"applied_at": r.AppliedAt,
			"raw_row":    json.RawMessage(r.RawRow),
		}

		b, _ := json.Marshal(payload)
		payloadJSON := string(b)

		_, exErr := tx.Exec(ctx, `
			INSERT INTO crm_outbox(crm_id, application_id, payload, status, attempt, created_at, updated_at)
			VALUES($1, $2, $3::jsonb, 'PENDING', 0, now(), now())
		`, uuid.New(), r.AppID, payloadJSON)
		if exErr != nil {
			res.Skipped++
			res.Errors = append(res.Errors, models.ActionItemError{
				ApplicationID: r.AppID.String(),
				Error:         fmt.Sprintf("не удалось добавить в crm_outbox: %v", exErr),
			})
			continue
		}

		_, exErr = tx.Exec(ctx, `
			UPDATE applications
			SET status=$2, updated_at=now()
			WHERE application_id=$1
		`, r.AppID, models.AppCRMQueued)
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

	err = tx.Commit(ctx)
	if err != nil {
		return models.BulkCRMActionResponse{}, err
	}
	return res, nil
}
