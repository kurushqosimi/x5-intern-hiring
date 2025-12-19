package models

type BulkEmailActionRequest struct {
	ApplicationIDs []string `json:"application_ids"`
	TemplateCode   string   `json:"template_code,omitempty"`
	StatusReason   string   `json:"status_reason,omitempty"` // актуально для reject
}

type BulkEmailActionResponse struct {
	Queued  int               `json:"queued"`
	Skipped int               `json:"skipped"`
	Errors  []ActionItemError `json:"errors,omitempty"`
}
