package repository

import (
	"go-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(role *models.Role) error
	FindAll(limit, offset int) ([]models.Role, int64, error)
	FindByID(id uuid.UUID) (*models.Role, error)
	FindActiveByDepartmentID(departmentID uuid.UUID, limit, offset int) ([]models.Role, int64, error)
	Update(role *models.Role) error
	Delete(id uuid.UUID) error
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) FindAll(limit, offset int) ([]models.Role, int64, error) {
	var roles []models.Role
	var totalCount int64

	err := r.db.Model(&models.Role{}).Where("status = ?", "active").Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Department").Where("status = ?", "active").Limit(limit).Offset(offset).Find(&roles).Error
	return roles, totalCount, err
}

func (r *roleRepository) FindActiveByDepartmentID(departmentID uuid.UUID, limit, offset int) ([]models.Role, int64, error) {
	var roles []models.Role
	var totalCount int64

	err := r.db.Model(&models.Role{}).Where("department_id = ? AND status = ?", departmentID, "active").Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Department").Where("department_id = ? AND status = ?", departmentID, "active").Limit(limit).Offset(offset).Find(&roles).Error
	return roles, totalCount, err
}

func (r *roleRepository) FindByID(id uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Department").First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Update(role *models.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Role{}, "id = ?", id).Error
}
