package models

import (
	"time"

	"github.com/google/uuid"
)

type EmployeeScreenshot struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EmployeeID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee         *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	EmployeeDeviceID *uuid.UUID `gorm:"type:uuid;index" json:"employee_device_id"`
	DepartmentID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"department_id"`
	Filename         string     `gorm:"type:varchar(255);not null" json:"filename"`
	ContentType      string     `gorm:"type:varchar(100);not null" json:"content_type"`
	FileSize         int64      `gorm:"not null" json:"file_size"`
	FilePath         string     `gorm:"type:text;not null" json:"file_path"`
	CapturedAt       time.Time  `gorm:"not null;index" json:"captured_at"`
	Notes            string     `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}
