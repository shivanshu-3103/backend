package models

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID                     uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID         uuid.UUID  `gorm:"type:uuid" json:"organization_id"`
	EmployeeCode           string     `gorm:"unique;not null" json:"employee_code"`
	Name                   string     `gorm:"not null" json:"name"`
	Email                  string     `gorm:"unique;not null" json:"email"`
	PasswordHash           string     `gorm:"not null" json:"-"`
	Phone                  string     `json:"phone"`
	DateOfJoining          *time.Time `json:"date_of_joining"`
	Designation            string     `json:"designation"`
	DepartmentID           uuid.UUID   `gorm:"type:uuid" json:"department_id"`
	Department             *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	RoleID                 uuid.UUID   `gorm:"type:uuid" json:"role_id"`
	Role                   *Role       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	TeamID                 *uuid.UUID  `gorm:"type:uuid" json:"team_id"`
	Team                   *Team       `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	ReportingManager       *uuid.UUID  `gorm:"type:uuid" json:"reporting_manager"`
	WorkLocation           string     `json:"work_location"`
	EmploymentType         string     `json:"employment_type"`
	Photo                  string     `json:"photo"`
	Address                string     `json:"address"`
	DateOfBirth            *time.Time `json:"date_of_birth"`
	Gender                 string     `json:"gender"`
	BloodGroup             string     `json:"blood_group"`
	EmergencyContactName   string     `json:"emergency_contact_name"`
	EmergencyContactNumber string     `json:"emergency_contact_number"`
	Relationship           string     `json:"relationship"`
	Notes                  string     `json:"notes"`
	Status                 string     `gorm:"default:'active'" json:"status"`
	LastActiveAt           time.Time  `json:"last_active_at"`
	CreatedAt              time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}
