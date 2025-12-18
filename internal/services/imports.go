package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/kurushqosimi/x5-intern-hiring/internal/custom_errors"
	"github.com/kurushqosimi/x5-intern-hiring/internal/models"
	"github.com/xuri/excelize/v2"
	"io"
	"mime/multipart"
	"strings"
	"time"
)

func (s *Service) ProcessXLSX(ctx context.Context, fileHeader *multipart.FileHeader) (*models.XLSXProcRes, error) {
	metadata, content, err := s.readFile(fileHeader)
	if err != nil {
		return nil, err
	}

	if err = s.repo.InsertImport(ctx, metadata); err != nil {
		return nil, err
	}

	parsedRows, parsedErrors, err := s.parseXLSX(content)
	if err != nil {
		return nil, err
	}
	inserted, skipped, err := s.repo.InsertXLSXData(ctx, metadata.ImportID, parsedRows)
	if err != nil {
		_ = s.repo.SetImportStats(ctx, metadata.ImportID, models.FileFailed, len(parsedRows), inserted, skipped)
		return nil, err
	}

	_ = s.repo.SetImportStats(ctx, metadata.ImportID, models.FileParsed, len(parsedRows), inserted, skipped)

	return &models.XLSXProcRes{
		ImportId:     metadata.ImportID.String(),
		FileSha256:   metadata.FileSha256,
		TotalRows:    len(parsedRows),
		InsertedRows: inserted,
		SkippedRows:  skipped,
		Errors:       parsedErrors,
	}, nil
}

func (s *Service) readFile(fileHeader *multipart.FileHeader) (*models.FileMetaData, []byte, error) {
	f, err := fileHeader.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("%v: %w", err, custom_errors.ErrFailedToOpenFile)
	}
	defer f.Close()

	// чтобы и sha256 посчитать и excelize прочитать — читаем в память (для больших файлов лучше в temp file)
	buf, err := io.ReadAll(io.LimitReader(f, 50<<20)) // 50MB лимит
	if err != nil {
		return nil, nil, fmt.Errorf("%v: %w", err, custom_errors.ErrFailedToReadFile)
	}

	sum := sha256.Sum256(buf)
	fileSHA := hex.EncodeToString(sum[:])

	return &models.FileMetaData{
		ImportID:     uuid.New(),
		UploadedBy:   uuid.Nil,
		FileName:     fileHeader.Filename,
		FileSha256:   fileSHA,
		Status:       models.FileCreated,
		TotalRows:    0,
		InsertedRows: 0,
		SkippedRows:  0,
	}, buf, nil
}

func (s *Service) parseXLSX(buf []byte) ([]models.ParsedRow, []string, error) {
	xl, err := excelize.OpenReader(bytes.NewReader(buf))
	if err != nil {
		return nil, nil, fmt.Errorf("%v: %w", err, custom_errors.ErrInvalidXLSX)
	}
	defer func() { _ = xl.Close() }()

	sheet := xl.GetSheetName(0)
	if sheet == "" {
		return nil, nil, fmt.Errorf("%v: %w", err, custom_errors.ErrNoXLSXSheets)
	}

	rows, err := xl.GetRows(sheet)
	if err != nil || len(rows) < 2 {
		return nil, nil, fmt.Errorf("%v: %w", err, custom_errors.ErrNoXLSXData)
	}

	header := rows[0]
	col := indexColumns(header)

	var parsed []models.ParsedRow
	var parseErrors []string

	for i := 1; i < len(rows); i++ {
		r := rows[i]
		// пустая строка
		if len(r) == 0 || allEmpty(r) {
			continue
		}

		get := func(name string) string {
			idx, ok := col[name]
			if !ok || idx >= len(r) {
				return ""
			}
			return strings.TrimSpace(r[idx])
		}

		lastName := get(models.LastName)
		firstName := get(models.FirstName)
		email := strings.ToLower(strings.TrimSpace(get(models.Email)))
		phone := normalizePhone(get(models.Cellphone))
		telegram := strings.TrimSpace(get(models.Telegram))

		if lastName == "" && firstName == "" {
			parseErrors = append(parseErrors, fmt.Sprintf("строка %d: пустое имя", i+1))
			continue
		}
		if email == "" && phone == "" {
			parseErrors = append(parseErrors, fmt.Sprintf("строка %d: имейл и номер телефона пусты", i+1))
			continue
		}

		appliedAt, err := parseAppliedAt(get(models.ApplicationDate))
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("строка %d: инвалидная дата подачи: %v", i+1, err))
			continue
		}

		var by *int
		if s := get(models.YearBorn); s != "" {
			v, e := parseYear(s)
			if e != nil {
				parseErrors = append(parseErrors, fmt.Sprintf("строка %d: инвалидная дата рождения: %v", i+1, e))
			} else {
				by = &v
			}
		}

		raw := map[string]any{}
		for j, name := range header {
			if name == "" || j >= len(r) {
				continue
			}
			raw[name] = r[j]
		}

		parsed = append(parsed, models.ParsedRow{
			LastName:        lastName,
			FirstName:       firstName,
			Email:           email,
			Phone:           phone,
			Telegram:        telegram,
			ResumeURL:       get(models.ResumeURL),
			Priority1:       get(models.FirstPriority),
			Priority2:       get(models.SecondPriority),
			Course:          get(models.Course),
			Specialty:       get(models.Speciality),
			SpecialtyOther:  get(models.OtherSpeciality),
			Schedule:        get(models.Schedule),
			City:            get(models.City),
			CityOther:       get(models.OtherCity),
			Source:          get(models.Source),
			BirthYear:       by,
			Citizenship:     get(models.Citizenship),
			University:      get(models.University),
			UniversityOther: get(models.OtherUniversity),
			Languages:       get(models.ProgrammingLanguages),
			AppliedAt:       appliedAt,
			RawRow:          raw,
		})
	}

	return parsed, parseErrors, nil
}

// indexColumns - взятие название столбцов в xlsx
func indexColumns(header []string) map[string]int {
	m := make(map[string]int, len(header))
	for i, h := range header {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		m[h] = i
	}
	return m
}

func allEmpty(r []string) bool {
	for _, s := range r {
		if strings.TrimSpace(s) != "" {
			return false
		}
	}
	return true
}

func normalizePhone(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// оставим только цифры и +
	var b strings.Builder
	for _, ch := range s {
		if (ch >= '0' && ch <= '9') || ch == '+' {
			b.WriteRune(ch)
		}
	}
	out := b.String()
	// если начинается с 8 и длина 11 (RU), например, можно привести к +7...
	return out
}

func parseAppliedAt(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("empty")
	}
	// Подстрой под реальный формат в файле.
	// Часто встречается: "02.01.2006 15:04" или "02.01.2006"
	layouts := []string{
		"02.01.2006 15:04:05",
		"02.01.2006 15:04",
		"02.01.2006",
		time.RFC3339,
	}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, s, time.Local); err == nil {
			// если дата без времени — ставим 00:00 локально
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("неподдерживаемый формат даты: %q", s)
}

func parseYear(s string) (int, error) {
	s = strings.TrimSpace(s)
	if len(s) == 4 {
		var y int
		_, err := fmt.Sscanf(s, "%d", &y)
		return y, err
	}
	// если в xlsx могло прийти типа "2001 г."
	var y int
	_, err := fmt.Sscanf(s, "%d", &y)
	return y, err
}
