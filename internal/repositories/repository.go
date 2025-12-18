package repositories

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kurushqosimi/x5-intern-hiring/internal/models"
	"strings"
	"time"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const (
	candidateInsert = `
	INSERT INTO candidates(candidate_id, first_name, last_name, birth_year, citizenship, languages)
	VALUES($1,$2,$3,$4,$5,$6)
`
	candidateEmailInsert = `
	INSERT INTO candidate_contacts(contact_id, candidate_id, type, value, is_primary, normalized)
	VALUES($1,$2,'email',$3,true,$4)
	ON CONFLICT (type, normalized) DO NOTHING
`
	candidatePhoneInsert = `
	INSERT INTO candidate_contacts(contact_id, candidate_id, type, value, is_primary, normalized)
	VALUES($1,$2,'phone',$3,true,$4)
	ON CONFLICT (type, normalized) DO NOTHING
`
	candidateTelegramInsert = `
	INSERT INTO candidate_contacts(contact_id, candidate_id, type, value, is_primary, normalized)
	VALUES($1,$2,'telegram',$3,true,$4)
	ON CONFLICT (type, normalized) DO NOTHING
`
	applicationsInsert = `
	INSERT INTO applications(
		application_id, candidate_id, import_id, applied_at, resume_url,
		priority1, priority2, course, specialty, specialty_other, schedule,
		city, city_other, university, university_other, source,
		status, status_reason, external_key, raw_row
	)
	VALUES(
		$1,$2,$3,$4,$5,
		$6,$7,$8,$9,$10,$11,
		$12,$13,$14,$15,$16,
		$17,$18,$19,$20
	)
	ON CONFLICT (external_key) DO NOTHING
`
)

// InsertXLSXData - вставка данных с файла в таблицы
func (repo *Repository) InsertXLSXData(
	ctx context.Context,
	importID uuid.UUID,
	rows []models.ParsedRow,
) (
	inserted int,
	skipped int,
	err error,
) {
	tx, err := repo.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	for _, r := range rows {
		// 1) candidate_id: здесь делаем просто новый candidate на каждую строку
		// (позже можно улучшить: искать кандидата по контакту и переиспользовать)
		candidateID := uuid.New()

		_, err = tx.Exec(ctx, candidateInsert, candidateID, r.FirstName, r.LastName, r.BirthYear, nullIfEmpty(r.Citizenship), nullIfEmpty(r.Languages))
		if err != nil {
			return inserted, skipped, err
		}

		// 2) contacts
		if r.Email != "" {
			_, _ = tx.Exec(ctx, candidateEmailInsert, uuid.New(), candidateID, r.Email, r.Email)
		}
		if r.Phone != "" {
			_, _ = tx.Exec(ctx, candidatePhoneInsert, uuid.New(), candidateID, r.Phone, r.Phone)
		}
		if r.Telegram != "" {
			norm := strings.ToLower(strings.TrimPrefix(r.Telegram, "@"))
			_, _ = tx.Exec(ctx, candidateTelegramInsert, uuid.New(), candidateID, r.Telegram, norm)
		}

		// 3) application external_key
		key := buildExternalKey(r.Email, r.Phone, r.AppliedAt, r.Priority1, r.Priority2)

		appID := uuid.New()
		ct, err := tx.Exec(ctx, applicationsInsert, appID, candidateID, importID, r.AppliedAt, nullIfEmpty(r.ResumeURL),
			nullIfEmpty(r.Priority1), nullIfEmpty(r.Priority2), nullIfEmpty(r.Course), nullIfEmpty(r.Specialty),
			nullIfEmpty(r.SpecialtyOther), nullIfEmpty(r.Schedule),
			nullIfEmpty(r.City), nullIfEmpty(r.CityOther), nullIfEmpty(r.University), nullIfEmpty(r.UniversityOther),
			nullIfEmpty(r.Source),
			"NEW", nil, key, r.RawRow,
		)
		if err != nil {
			return inserted, skipped, err
		}
		if ct.RowsAffected() == 0 {
			skipped++
		} else {
			inserted++
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return inserted, skipped, err
	}
	return inserted, skipped, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func buildExternalKey(email, phone string, appliedAt time.Time, p1, p2 string) string {
	base := strings.ToLower(strings.TrimSpace(email))
	if base == "" {
		base = strings.TrimSpace(phone)
	}
	payload := fmt.Sprintf("%s|%s|%s|%s",
		base,
		appliedAt.UTC().Format(time.RFC3339),
		strings.TrimSpace(p1),
		strings.TrimSpace(p2),
	)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
