package repository

import (
	"go-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EmployeeDeviceRepository interface {
	UpsertDevice(device *models.EmployeeDevice) error
	FindAll(limit, offset int) ([]models.EmployeeDevice, int64, error)
	FindByEmployeeID(employeeID uuid.UUID, limit, offset int) ([]models.EmployeeDevice, int64, error)
	FindByID(id uuid.UUID) (*models.EmployeeDevice, error)
	FindByFingerprint(fingerprint string) (*models.EmployeeDevice, error)
	DeleteDevice(id uuid.UUID) error
}

type employeeDeviceRepository struct {
	db *gorm.DB
}

func NewEmployeeDeviceRepository(db *gorm.DB) EmployeeDeviceRepository {
	return &employeeDeviceRepository{db: db}
}

func (r *employeeDeviceRepository) UpsertDevice(device *models.EmployeeDevice) error {
	// Upsert based on unique constraint (device_fingerprint)
	// If the fingerprint exists, it will update the other provided columns
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "device_fingerprint"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"machine_guid_hash", "bios_serial_hash", "motherboard_serial_hash",
			"device_name", "manufacturer", "model", "os_name", "os_version",
			"architecture", "cpu_name", "ram_total_mb", "disk_total_gb",
			"app_version", "last_seen_at", "updated_at",
		}),
	}).Create(device).Error
}

func (r *employeeDeviceRepository) FindAll(limit, offset int) ([]models.EmployeeDevice, int64, error) {
	var devices []models.EmployeeDevice
	var totalCount int64

	if err := r.db.Model(&models.EmployeeDevice{}).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("Employee").Limit(limit).Offset(offset).Find(&devices).Error
	return devices, totalCount, err
}

func (r *employeeDeviceRepository) FindByEmployeeID(employeeID uuid.UUID, limit, offset int) ([]models.EmployeeDevice, int64, error) {
	var devices []models.EmployeeDevice
	var totalCount int64

	if err := r.db.Model(&models.EmployeeDevice{}).Where("employee_id = ?", employeeID).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("Employee").Where("employee_id = ?", employeeID).Limit(limit).Offset(offset).Find(&devices).Error
	return devices, totalCount, err
}

func (r *employeeDeviceRepository) FindByID(id uuid.UUID) (*models.EmployeeDevice, error) {
	var device models.EmployeeDevice
	err := r.db.First(&device, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *employeeDeviceRepository) FindByFingerprint(fingerprint string) (*models.EmployeeDevice, error) {
	var device models.EmployeeDevice
	err := r.db.Preload("Employee").First(&device, "device_fingerprint = ?", fingerprint).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *employeeDeviceRepository) DeleteDevice(id uuid.UUID) error {
	return r.db.Delete(&models.EmployeeDevice{}, "id = ?", id).Error
}
