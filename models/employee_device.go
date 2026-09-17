package models

import (
	"time"

	"github.com/google/uuid"
)

type EmployeeDevice struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EmployeeID            uuid.UUID  `gorm:"type:uuid;not null;index:idx_employee_devices_employee_id" json:"employee_id"`
	Employee              *Employee  `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE;" json:"employee,omitempty"`
	DeviceFingerprint     string     `gorm:"type:varchar(128);unique;not null" json:"device_fingerprint"`
	MachineGuidHash       string     `gorm:"type:varchar(128)" json:"machine_guid_hash"`
	BiosSerialHash        string     `gorm:"type:varchar(128)" json:"bios_serial_hash"`
	MotherboardSerialHash string     `gorm:"type:varchar(128)" json:"motherboard_serial_hash"`
	DeviceName            string     `gorm:"type:varchar(255)" json:"device_name"`
	Manufacturer          string     `gorm:"type:varchar(255)" json:"manufacturer"`
	Model                 string     `gorm:"type:varchar(255)" json:"model"`
	OSName                string     `gorm:"type:varchar(100)" json:"os_name"`
	OSVersion             string     `gorm:"type:varchar(100)" json:"os_version"`
	Architecture          string     `gorm:"type:varchar(20)" json:"architecture"`
	CPUName               string     `gorm:"type:varchar(255)" json:"cpu_name"`
	RamTotalMB            int64      `json:"ram_total_mb"`
	DiskTotalGB           int64      `json:"disk_total_gb"`
	AppVersion            string     `gorm:"type:varchar(50)" json:"app_version"`
	Status                string     `gorm:"type:varchar(30);not null;default:'active';index" json:"status"`
	LastSeenAt            *time.Time `gorm:"index" json:"last_seen_at"`
	RegisteredAt          time.Time  `gorm:"autoCreateTime" json:"registered_at"`
	UpdatedAt             time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}
