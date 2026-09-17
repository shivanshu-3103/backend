package repository

import (
	"go-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeScreenshotRepository interface {
	Create(*models.EmployeeScreenshot) error
	FindByID(uuid.UUID) (*models.EmployeeScreenshot, error)
	FindByEmployeeID(uuid.UUID, int, int) ([]models.EmployeeScreenshot, int64, error)
	FindAll(employeeID, departmentID *uuid.UUID, limit, offset int) ([]models.EmployeeScreenshot, int64, error)
	Update(*models.EmployeeScreenshot) error
	Delete(uuid.UUID) error
}

type employeeScreenshotRepository struct{ db *gorm.DB }

func NewEmployeeScreenshotRepository(db *gorm.DB) EmployeeScreenshotRepository {
	return &employeeScreenshotRepository{db: db}
}

func (r *employeeScreenshotRepository) Create(screenshot *models.EmployeeScreenshot) error {
	return r.db.Create(screenshot).Error
}

func (r *employeeScreenshotRepository) FindByID(id uuid.UUID) (*models.EmployeeScreenshot, error) {
	var screenshot models.EmployeeScreenshot
	err := r.db.Preload("Employee").First(&screenshot, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &screenshot, nil
}

func (r *employeeScreenshotRepository) FindByEmployeeID(employeeID uuid.UUID, limit, offset int) ([]models.EmployeeScreenshot, int64, error) {
	return r.find(&employeeID, nil, limit, offset)
}

func (r *employeeScreenshotRepository) FindAll(employeeID, departmentID *uuid.UUID, limit, offset int) ([]models.EmployeeScreenshot, int64, error) {
	return r.find(employeeID, departmentID, limit, offset)
}

func (r *employeeScreenshotRepository) find(employeeID, departmentID *uuid.UUID, limit, offset int) ([]models.EmployeeScreenshot, int64, error) {
	var screenshots []models.EmployeeScreenshot
	var totalCount int64

	countQuery := r.db.Model(&models.EmployeeScreenshot{})
	if employeeID != nil {
		countQuery = countQuery.Where("employee_id = ?", *employeeID)
	}
	if departmentID != nil {
		countQuery = countQuery.Where("department_id = ?", *departmentID)
	}
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	query := r.db.Preload("Employee").Order("captured_at DESC")
	if employeeID != nil {
		query = query.Where("employee_id = ?", *employeeID)
	}
	if departmentID != nil {
		query = query.Where("department_id = ?", *departmentID)
	}
	err := query.Limit(limit).Offset(offset).Find(&screenshots).Error
	return screenshots, totalCount, err
}

func (r *employeeScreenshotRepository) Update(screenshot *models.EmployeeScreenshot) error {
	return r.db.Model(&models.EmployeeScreenshot{}).Where("id = ?", screenshot.ID).Updates(map[string]interface{}{
		"filename": screenshot.Filename, "content_type": screenshot.ContentType,
		"file_size": screenshot.FileSize, "file_path": screenshot.FilePath,
		"captured_at": screenshot.CapturedAt, "notes": screenshot.Notes,
	}).Error
}

func (r *employeeScreenshotRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.EmployeeScreenshot{}, "id = ?", id).Error
}
