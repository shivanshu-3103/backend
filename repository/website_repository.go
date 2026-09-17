package repository

import (
	"go-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WebsiteRepository interface {
	Create(website *models.Website) error
	FindByID(id uuid.UUID) (*models.Website, error)
	FindByEmployeeID(employeeID uuid.UUID, limit, offset int) ([]models.Website, int64, error)
	FindAll(employeeID, departmentID *uuid.UUID, limit, offset int) ([]models.Website, int64, error)
	Update(website *models.Website) error
	Delete(id uuid.UUID) error
}

type websiteRepository struct {
	db *gorm.DB
}

func NewWebsiteRepository(db *gorm.DB) WebsiteRepository {
	return &websiteRepository{db: db}
}

func (r *websiteRepository) Create(website *models.Website) error {
	return r.db.Create(website).Error
}

func (r *websiteRepository) FindByID(id uuid.UUID) (*models.Website, error) {
	var website models.Website
	err := r.db.Preload("Employee").First(&website, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &website, nil
}

func (r *websiteRepository) FindByEmployeeID(employeeID uuid.UUID, limit, offset int) ([]models.Website, int64, error) {
	return r.find(&employeeID, nil, limit, offset)
}

func (r *websiteRepository) FindAll(employeeID, departmentID *uuid.UUID, limit, offset int) ([]models.Website, int64, error) {
	return r.find(employeeID, departmentID, limit, offset)
}

func (r *websiteRepository) find(employeeID, departmentID *uuid.UUID, limit, offset int) ([]models.Website, int64, error) {
	var websites []models.Website
	var totalCount int64

	countQuery := r.db.Model(&models.Website{})
	if employeeID != nil {
		countQuery = countQuery.Where("websites.employee_id = ?", *employeeID)
	}
	if departmentID != nil {
		countQuery = countQuery.Joins("JOIN employees ON employees.id = websites.employee_id").
			Where("employees.department_id = ?", *departmentID)
	}
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	query := r.db.Preload("Employee").Order("started_at DESC")
	
	if employeeID != nil {
		query = query.Where("websites.employee_id = ?", *employeeID)
	}
	if departmentID != nil {
		query = query.Joins("JOIN employees ON employees.id = websites.employee_id").
			Where("employees.department_id = ?", *departmentID)
	}
	
	err := query.Limit(limit).Offset(offset).Find(&websites).Error
	return websites, totalCount, err
}

func (r *websiteRepository) Update(website *models.Website) error {
	return r.db.Save(website).Error
}

func (r *websiteRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Website{}, "id = ?", id).Error
}
