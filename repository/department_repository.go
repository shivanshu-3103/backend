package repository

import (
	"go-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DepartmentRepository interface {
	Create(department *models.Department) error
	FindAll(limit, offset int) ([]models.Department, int64, error)
	FindByID(id uuid.UUID) (*models.Department, error)
	Update(department *models.Department) error
	Delete(id uuid.UUID) error
}

type departmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) Create(department *models.Department) error {
	return r.db.Create(department).Error
}

func (r *departmentRepository) FindAll(limit, offset int) ([]models.Department, int64, error) {
	var departments []models.Department
	var totalCount int64

	err := r.db.Model(&models.Department{}).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Limit(limit).Offset(offset).Find(&departments).Error
	return departments, totalCount, err
}

func (r *departmentRepository) FindByID(id uuid.UUID) (*models.Department, error) {
	var department models.Department
	err := r.db.First(&department, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &department, nil
}

func (r *departmentRepository) Update(department *models.Department) error {
	return r.db.Save(department).Error
}

func (r *departmentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Department{}, "id = ?", id).Error
}
