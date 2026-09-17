package repository

import (
	"go-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApplicationRepository interface {
	Create(application *models.Application) error
	FindByID(id uuid.UUID) (*models.Application, error)
	FindByEmployeeID(employeeID uuid.UUID, limit, offset int) ([]models.Application, int64, error)
	FindAll(employeeID, departmentID *uuid.UUID, limit, offset int) ([]models.Application, int64, error)
	Update(application *models.Application) error
	Delete(id uuid.UUID) error
}

type applicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepository{db: db}
}

func (r *applicationRepository) Create(application *models.Application) error {
	return r.db.Create(application).Error
}

func (r *applicationRepository) FindByID(id uuid.UUID) (*models.Application, error) {
	var application models.Application
	err := r.db.Preload("Employee").First(&application, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &application, nil
}

func (r *applicationRepository) FindByEmployeeID(employeeID uuid.UUID, limit, offset int) ([]models.Application, int64, error) {
	return r.find(&employeeID, nil, limit, offset)
}

func (r *applicationRepository) FindAll(employeeID, departmentID *uuid.UUID, limit, offset int) ([]models.Application, int64, error) {
	return r.find(employeeID, departmentID, limit, offset)
}

func (r *applicationRepository) find(employeeID, departmentID *uuid.UUID, limit, offset int) ([]models.Application, int64, error) {
	var applications []models.Application
	var totalCount int64

	countQuery := r.db.Model(&models.Application{})
	if employeeID != nil {
		countQuery = countQuery.Where("applications.employee_id = ?", *employeeID)
	}
	if departmentID != nil {
		countQuery = countQuery.Joins("JOIN employees ON employees.id = applications.employee_id").
			Where("employees.department_id = ?", *departmentID)
	}
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	query := r.db.Preload("Employee").Order("started_at DESC")
	
	if employeeID != nil {
		query = query.Where("applications.employee_id = ?", *employeeID)
	}
	if departmentID != nil {
		query = query.Joins("JOIN employees ON employees.id = applications.employee_id").
			Where("employees.department_id = ?", *departmentID)
	}
	
	err := query.Limit(limit).Offset(offset).Find(&applications).Error
	return applications, totalCount, err
}

func (r *applicationRepository) Update(application *models.Application) error {
	return r.db.Save(application).Error
}

func (r *applicationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Application{}, "id = ?", id).Error
}
