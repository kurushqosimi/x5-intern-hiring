package models

type XLSXProcRes struct {
	ImportId     string   `json:"import_id"`
	FileSha256   string   `json:"file_sha256"`
	TotalRows    int      `json:"total_rows"`
	InsertedRows int      `json:"inserted_rows"`
	SkippedRows  int      `json:"skipped_rows"`
	Errors       []string `json:"errors"`
}
