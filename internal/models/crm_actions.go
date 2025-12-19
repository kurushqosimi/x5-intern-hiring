package models

type BulkCRMActionRequest struct {
	ApplicationIDs []string `json:"application_ids"`
}

type BulkCRMActionResponse struct {
	Queued  int               `json:"queued"`
	Skipped int               `json:"skipped"`
	Errors  []ActionItemError `json:"errors,omitempty"`
}
