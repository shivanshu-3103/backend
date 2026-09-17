package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Task struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name              string         `gorm:"type:varchar(255);not null" json:"name"`
	Type              string         `gorm:"type:varchar(50);not null" json:"type"` // e.g., "work", "break", "other"
	IdleLogoutEnabled bool           `gorm:"default:false" json:"idle_logout_enabled"`
	IdleLogoutMinutes *int           `json:"idle_logout_minutes"`
	Status            string         `gorm:"type:varchar(50);default:'active'" json:"status"`
	CreatedBy         *int           `gorm:"type:integer" json:"created_by"`
	Creator           *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
