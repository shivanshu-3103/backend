package models

import (
	"time"

	"github.com/google/uuid"
)

type Attendance struct {
	ID             uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EmployeeID     uuid.UUID         `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee       *Employee         `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Date           time.Time         `gorm:"type:date;not null;index" json:"date"`
	FirstClockIn   *time.Time        `json:"first_clock_in"`
	LastClockOut   *time.Time        `json:"last_clock_out"`
	WorkType       string            `json:"work_type"` // office, remote
	Status         string            `json:"status"`    // working, on_break, completed
	TotalWorkSecs  int               `json:"total_work_secs"`
	TotalBreakSecs int               `json:"total_break_secs"`
	CreatedAt      time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
	Events         []AttendanceEvent `gorm:"foreignKey:AttendanceID" json:"events,omitempty"`
}

type AttendanceEvent struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AttendanceID uuid.UUID   `gorm:"type:uuid;not null;index" json:"attendance_id"`
	EventType    string      `json:"event_type"` // clock_in, clock_out, break_start, break_end
	EventTime    time.Time   `json:"event_time"`
	Reason       string      `json:"reason,omitempty"`       // e.g., Lunch, Tea, Other
	OtherReason  string      `json:"other_reason,omitempty"` // For custom reasons when Reason is 'Other'
	CreatedAt    time.Time   `gorm:"autoCreateTime" json:"created_at"`
}
