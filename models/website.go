package models

import (
	"time"

	"github.com/google/uuid"
)

type Website struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EmployeeID      uuid.UUID `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee        *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Domain          string    `gorm:"not null" json:"domain"`
	URL             string    `gorm:"not null" json:"url"`
	PageTitle       string    `json:"page_title"`
	StartedAt       time.Time `gorm:"not null" json:"started_at"`
	EndedAt         time.Time `json:"ended_at"`
	DurationSeconds int64     `gorm:"default:0" json:"duration_seconds"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
