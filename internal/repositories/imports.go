package repositories

import (
	"context"
	"github.com/google/uuid"
	"github.com/kurushqosimi/x5-intern-hiring/internal/models"
)

// InsertImport - вставка метаданных об документе
func (repo *Repository) InsertImport(ctx context.Context, fileMetadata *models.FileMetaData) error {
	const query = `
		INSERT INTO imports(import_id, uploaded_by, file_name, file_sha256, status, total_rows, inserted_rows, skipped_rows)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := repo.pool.Exec(
		ctx, query, fileMetadata.ImportID, nil, fileMetadata.FileName, fileMetadata.FileSha256, fileMetadata.Status,
		fileMetadata.TotalRows, fileMetadata.InsertedRows, fileMetadata.SkippedRows)

	return err
}

func (repo *Repository) SetImportFailed(ctx context.Context, importID uuid.UUID, status string) error {
	const query = `UPDATE imports SET status=$2 WHERE import_id=$1`
	_, err := repo.pool.Exec(ctx, query, importID, status)
	return err
}

func (repo *Repository) SetImportStats(ctx context.Context, importID uuid.UUID, status string, total, inserted, skipped int) error {
	const query = `
		UPDATE imports
		SET status=$2, total_rows=$3, inserted_rows=$4, skipped_rows=$5
		WHERE import_id=$1
	`
	_, err := repo.pool.Exec(ctx, query, importID, status, total, inserted, skipped)
	return err
}
