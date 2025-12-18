package models

import "time"

type ListApplicationsParams struct {
	Limit  int
	Offset int

	Statuses []string

	Q string

	Priority    string
	Course      string
	Specialty   string
	Schedule    string
	City        string
	University  string
	Citizenship string

	AppliedFrom *time.Time
	AppliedTo   *time.Time

	HasResume *bool
	ImportID  string // optional (uuid as string)
}

type ApplicationListItem struct {
	ApplicationID string    `json:"application_id"`
	CandidateID   string    `json:"candidate_id"`
	ImportID      string    `json:"import_id"`
	AppliedAt     time.Time `json:"applied_at"`
	Status        string    `json:"status"`

	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	BirthYear   *int   `json:"birth_year,omitempty"`
	Citizenship string `json:"citizenship,omitempty"`
	Languages   string `json:"languages,omitempty"`

	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Telegram string `json:"telegram,omitempty"`

	ResumeURL string `json:"resume_url,omitempty"`

	Priority1 string `json:"priority1,omitempty"`
	Priority2 string `json:"priority2,omitempty"`
	Course    string `json:"course,omitempty"`

	Specialty      string `json:"specialty,omitempty"`
	SpecialtyOther string `json:"specialty_other,omitempty"`

	Schedule string `json:"schedule,omitempty"`

	City      string `json:"city,omitempty"`
	CityOther string `json:"city_other,omitempty"`

	University      string `json:"university,omitempty"`
	UniversityOther string `json:"university_other,omitempty"`

	Source string `json:"source,omitempty"`

	StatusReason string `json:"status_reason,omitempty"`
}

type ListApplicationsResponse struct {
	Items  []ApplicationListItem `json:"items"`
	Total  int                   `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}
