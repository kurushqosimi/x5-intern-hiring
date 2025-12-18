package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/kurushqosimi/x5-intern-hiring/internal/models"
	"strings"
)

func (repo *Repository) ListApplications(ctx context.Context, p models.ListApplicationsParams) (items []models.ApplicationListItem, total int, err error) {
	conds := []string{"1=1"}
	args := make([]any, 0, 16)

	addArg := func(v any) int {
		args = append(args, v)
		return len(args)
	}

	// status filter (comma-separated -> []string)
	if len(p.Statuses) > 0 {
		i := addArg(p.Statuses)
		conds = append(conds, fmt.Sprintf("a.status = ANY($%d::text[])", i))
	}

	// q search by name or any contact
	if q := strings.TrimSpace(p.Q); q != "" {
		q = "%" + q + "%"
		i := addArg(q)
		conds = append(conds, fmt.Sprintf(`
			(
				c.first_name ILIKE $%d OR
				c.last_name  ILIKE $%d OR
				EXISTS (
					SELECT 1 FROM candidate_contacts cc
					WHERE cc.candidate_id = a.candidate_id
					  AND cc.value ILIKE $%d
				)
			)
		`, i, i, i))
	}

	like2 := func(val string, expr string) {
		val = strings.TrimSpace(val)
		if val == "" {
			return
		}
		i := addArg("%" + val + "%")
		conds = append(conds, fmt.Sprintf(expr, i))
	}

	eq := func(val string, expr string) {
		val = strings.TrimSpace(val)
		if val == "" {
			return
		}
		i := addArg(val)
		conds = append(conds, fmt.Sprintf(expr, i))
	}

	// filters
	like2(p.Priority, "(a.priority1 ILIKE $%d OR a.priority2 ILIKE $%d)")
	if strings.TrimSpace(p.Priority) != "" {
		// like2 above добавит один arg, но нам нужно два плейсхолдера на один и тот же arg
		// => обработаем корректно вручную:
		args = args[:len(args)-1]
		i := addArg("%" + strings.TrimSpace(p.Priority) + "%")
		conds[len(conds)-1] = fmt.Sprintf("(a.priority1 ILIKE $%d OR a.priority2 ILIKE $%d)", i, i)
	}

	like2(p.Course, "a.course ILIKE $%d")
	like2(p.Specialty, "(a.specialty ILIKE $%d OR a.specialty_other ILIKE $%d)")
	if strings.TrimSpace(p.Specialty) != "" {
		args = args[:len(args)-1]
		i := addArg("%" + strings.TrimSpace(p.Specialty) + "%")
		conds[len(conds)-1] = fmt.Sprintf("(a.specialty ILIKE $%d OR a.specialty_other ILIKE $%d)", i, i)
	}

	like2(p.Schedule, "a.schedule ILIKE $%d")
	like2(p.City, "(a.city ILIKE $%d OR a.city_other ILIKE $%d)")
	if strings.TrimSpace(p.City) != "" {
		args = args[:len(args)-1]
		i := addArg("%" + strings.TrimSpace(p.City) + "%")
		conds[len(conds)-1] = fmt.Sprintf("(a.city ILIKE $%d OR a.city_other ILIKE $%d)", i, i)
	}

	like2(p.University, "(a.university ILIKE $%d OR a.university_other ILIKE $%d)")
	if strings.TrimSpace(p.University) != "" {
		args = args[:len(args)-1]
		i := addArg("%" + strings.TrimSpace(p.University) + "%")
		conds[len(conds)-1] = fmt.Sprintf("(a.university ILIKE $%d OR a.university_other ILIKE $%d)", i, i)
	}

	eq(p.Citizenship, "c.citizenship = $%d")

	if p.AppliedFrom != nil {
		i := addArg(*p.AppliedFrom)
		conds = append(conds, fmt.Sprintf("a.applied_at >= $%d", i))
	}
	if p.AppliedTo != nil {
		i := addArg(*p.AppliedTo)
		conds = append(conds, fmt.Sprintf("a.applied_at <= $%d", i))
	}

	if p.HasResume != nil {
		if *p.HasResume {
			conds = append(conds, "(a.resume_url IS NOT NULL AND a.resume_url <> '')")
		} else {
			conds = append(conds, "(a.resume_url IS NULL OR a.resume_url = '')")
		}
	}

	if strings.TrimSpace(p.ImportID) != "" {
		i := addArg(p.ImportID)
		conds = append(conds, fmt.Sprintf("a.import_id::text = $%d", i))
	}

	// limit/offset
	lim := p.Limit
	off := p.Offset
	if lim <= 0 {
		lim = 50
	}
	if lim > 200 {
		lim = 200
	}
	if off < 0 {
		off = 0
	}
	iLim := addArg(lim)
	iOff := addArg(off)

	where := strings.Join(conds, " AND ")

	qry := fmt.Sprintf(`
		SELECT
			a.application_id::text,
			a.candidate_id::text,
			a.import_id::text,
			a.applied_at,
			a.status,

			c.first_name,
			c.last_name,
			c.birth_year,
			c.citizenship,
			c.languages,

			(SELECT cc.value FROM candidate_contacts cc
			  WHERE cc.candidate_id=a.candidate_id AND cc.type='email'
			  ORDER BY cc.is_primary DESC, cc.created_at DESC
			  LIMIT 1) AS email,

			(SELECT cc.value FROM candidate_contacts cc
			  WHERE cc.candidate_id=a.candidate_id AND cc.type='phone'
			  ORDER BY cc.is_primary DESC, cc.created_at DESC
			  LIMIT 1) AS phone,

			(SELECT cc.value FROM candidate_contacts cc
			  WHERE cc.candidate_id=a.candidate_id AND cc.type='telegram'
			  ORDER BY cc.is_primary DESC, cc.created_at DESC
			  LIMIT 1) AS telegram,

			a.resume_url,
			a.priority1,
			a.priority2,
			a.course,
			a.specialty,
			a.specialty_other,
			a.schedule,
			a.city,
			a.city_other,
			a.university,
			a.university_other,
			a.source,
			a.status_reason,

			COUNT(*) OVER() AS total
		FROM applications a
		JOIN candidates c ON c.candidate_id = a.candidate_id
		WHERE %s
		ORDER BY a.applied_at DESC, a.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, iLim, iOff)

	rows, err := repo.pool.Query(ctx, qry, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var lastTotal int64 = 0

	for rows.Next() {
		var it models.ApplicationListItem

		var birth sql.NullInt32
		var citizenship sql.NullString
		var langs sql.NullString

		var email sql.NullString
		var phone sql.NullString
		var tg sql.NullString

		var resume sql.NullString
		var p1 sql.NullString
		var p2 sql.NullString
		var course sql.NullString
		var spec sql.NullString
		var specO sql.NullString
		var sched sql.NullString
		var city sql.NullString
		var cityO sql.NullString
		var uni sql.NullString
		var uniO sql.NullString
		var src sql.NullString
		var reason sql.NullString

		var totalRow int64

		err = rows.Scan(
			&it.ApplicationID,
			&it.CandidateID,
			&it.ImportID,
			&it.AppliedAt,
			&it.Status,

			&it.FirstName,
			&it.LastName,
			&birth,
			&citizenship,
			&langs,

			&email,
			&phone,
			&tg,

			&resume,
			&p1,
			&p2,
			&course,
			&spec,
			&specO,
			&sched,
			&city,
			&cityO,
			&uni,
			&uniO,
			&src,
			&reason,

			&totalRow,
		)
		if err != nil {
			return nil, 0, err
		}

		if birth.Valid {
			v := int(birth.Int32)
			it.BirthYear = &v
		}
		if citizenship.Valid {
			it.Citizenship = citizenship.String
		}
		if langs.Valid {
			it.Languages = langs.String
		}
		if email.Valid {
			it.Email = email.String
		}
		if phone.Valid {
			it.Phone = phone.String
		}
		if tg.Valid {
			it.Telegram = tg.String
		}
		if resume.Valid {
			it.ResumeURL = resume.String
		}
		if p1.Valid {
			it.Priority1 = p1.String
		}
		if p2.Valid {
			it.Priority2 = p2.String
		}
		if course.Valid {
			it.Course = course.String
		}
		if spec.Valid {
			it.Specialty = spec.String
		}
		if specO.Valid {
			it.SpecialtyOther = specO.String
		}
		if sched.Valid {
			it.Schedule = sched.String
		}
		if city.Valid {
			it.City = city.String
		}
		if cityO.Valid {
			it.CityOther = cityO.String
		}
		if uni.Valid {
			it.University = uni.String
		}
		if uniO.Valid {
			it.UniversityOther = uniO.String
		}
		if src.Valid {
			it.Source = src.String
		}
		if reason.Valid {
			it.StatusReason = reason.String
		}

		lastTotal = totalRow
		items = append(items, it)
	}

	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}

	return items, int(lastTotal), nil
}
