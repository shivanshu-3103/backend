package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EmployeeActivity struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EmployeeID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee         *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	EmployeeDeviceID *uuid.UUID `gorm:"type:uuid;index" json:"employee_device_id,omitempty"`
	KeyboardActivity int64      `gorm:"not null;default:0" json:"keyboard_activity"`
	MouseActivity    int64      `gorm:"not null;default:0" json:"mouse_activity"`
	Keystrokes       json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"keystrokes"`
	MouseEvents      json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"mouse_events"`
	CapturedAt       time.Time       `gorm:"not null;index" json:"captured_at"`
	CreatedAt        time.Time       `gorm:"autoCreateTime" json:"created_at"`
}
