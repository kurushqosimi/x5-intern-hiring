package models

import (
	"github.com/google/uuid"
	"time"
)

// file statuses
const (
	FileCreated = "CREATED"
	FileFailed  = "FAILED"
	FileParsed  = "PARSED"
)

// file columns' names
const (
	LastName             = "Фамилия"
	FirstName            = "Имя"
	Telegram             = "ТГ"
	Cellphone            = "Телефон"
	Email                = "Почта"
	ResumeURL            = "Резюме"
	FirstPriority        = "Первый приоритет"
	SecondPriority       = "Второй приоритет"
	Course               = "Курс"
	Speciality           = "Специальность"
	OtherSpeciality      = "Другая специальность"
	Schedule             = "График"
	City                 = "Город"
	OtherCity            = "Другой город"
	Source               = "Откуда узнал"
	YearBorn             = "Год рождения"
	Citizenship          = "Гражданство"
	University           = "ВУЗ"
	OtherUniversity      = "Другой ВУЗ"
	ProgrammingLanguages = "Языки"
	ApplicationDate      = "Дата заявки"
)

type ParsedRow struct {
	LastName        string
	FirstName       string
	Email           string
	Phone           string
	Telegram        string
	ResumeURL       string
	Priority1       string
	Priority2       string
	Course          string
	Specialty       string
	SpecialtyOther  string
	Schedule        string
	City            string
	CityOther       string
	Source          string
	BirthYear       *int
	Citizenship     string
	University      string
	UniversityOther string
	Languages       string
	AppliedAt       time.Time
	RawRow          map[string]any
}

type FileMetaData struct {
	ImportID     uuid.UUID
	UploadedBy   uuid.UUID
	FileName     string
	FileSha256   string
	Status       string
	TotalRows    int
	InsertedRows int
	SkippedRows  int
}
