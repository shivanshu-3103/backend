package models

import (
	"time"

	"github.com/google/uuid"
)

type TaskAssignment struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TaskID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"task_id"`
	Task       *Task      `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	AssignType string     `gorm:"type:varchar(50);not null" json:"assign_type"` // "company", "team", "employee"
	AssignID   *uuid.UUID `gorm:"type:uuid;index" json:"assign_id"`             // Null for company, UUID for team/employee
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}
