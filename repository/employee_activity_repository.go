package repository

import (
	"go-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeActivityRepository interface {
	Create(activity *models.EmployeeActivity) error
	FindByEmployeeID(employeeID uuid.UUID, limit, offset int) ([]models.EmployeeActivity, int64, error)
	FindAll(employeeID, departmentID *uuid.UUID, limit, offset int) ([]models.EmployeeActivity, int64, error)
}

type employeeActivityRepository struct {
	db *gorm.DB
}

func NewEmployeeActivityRepository(db *gorm.DB) EmployeeActivityRepository {
	return &employeeActivityRepository{db: db}
}

func (r *employeeActivityRepository) Create(activity *models.EmployeeActivity) error {
	return r.db.Create(activity).Error
}

func (r *employeeActivityRepository) FindByEmployeeID(employeeID uuid.UUID, limit, offset int) ([]models.EmployeeActivity, int64, error) {
	return r.find(&employeeID, nil, limit, offset)
}

func (r *employeeActivityRepository) FindAll(employeeID, departmentID *uuid.UUID, limit, offset int) ([]models.EmployeeActivity, int64, error) {
	return r.find(employeeID, departmentID, limit, offset)
}

func (r *employeeActivityRepository) find(employeeID, departmentID *uuid.UUID, limit, offset int) ([]models.EmployeeActivity, int64, error) {
	var activities []models.EmployeeActivity
	var totalCount int64

	countQuery := r.db.Model(&models.EmployeeActivity{})
	if employeeID != nil {
		countQuery = countQuery.Where("employee_activities.employee_id = ?", *employeeID)
	}
	if departmentID != nil {
		countQuery = countQuery.Joins("JOIN employees ON employees.id = employee_activities.employee_id").
			Where("employees.department_id = ?", *departmentID)
	}
	
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	query := r.db.Preload("Employee").Order("captured_at DESC")
	if employeeID != nil {
		query = query.Where("employee_activities.employee_id = ?", *employeeID)
	}
	if departmentID != nil {
		query = query.Joins("JOIN employees ON employees.id = employee_activities.employee_id").
			Where("employees.department_id = ?", *departmentID)
	}
	err := query.Limit(limit).Offset(offset).Find(&activities).Error
	return activities, totalCount, err
}
