package repository

import (
	"fmt"
	"go-api/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeRepository interface {
	CreateWithUser(employee *models.Employee, user *models.User) error
	FindAll(limit, offset int) ([]models.Employee, int64, error)
	FindByID(id uuid.UUID) (*models.Employee, error)
	UpdateWithUser(employee *models.Employee) error
	DeleteWithUser(id uuid.UUID) error
}

type employeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) CreateWithUser(employee *models.Employee, user *models.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create login record: %w", err)
		}
		if err := tx.Create(employee).Error; err != nil {
			return fmt.Errorf("failed to create employee record: %w", err)
		}
		return nil
	})
}

func (r *employeeRepository) FindAll(limit, offset int) ([]models.Employee, int64, error) {
	var employees []models.Employee
	var totalCount int64

	// Count total records
	if err := r.db.Model(&models.Employee{}).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated and preloaded records
	err := r.db.Preload("Department").Preload("Role").Preload("Team").Limit(limit).Offset(offset).Find(&employees).Error
	return employees, totalCount, err
}

func (r *employeeRepository) FindByID(id uuid.UUID) (*models.Employee, error) {
	var employee models.Employee
	err := r.db.Preload("Department").Preload("Role").Preload("Team").First(&employee, id).Error
	return &employee, err
}

func (r *employeeRepository) UpdateWithUser(employee *models.Employee) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(employee).Error; err != nil {
			return fmt.Errorf("failed to update employee: %w", err)
		}
		// Sync name and email to the users table
		if err := tx.Model(&models.User{}).Where("email = ?", employee.Email).Updates(map[string]interface{}{
			"name":  employee.Name,
			"email": employee.Email,
		}).Error; err != nil {
			// Not a fatal error — user record may not exist yet for older employees
		}
		return nil
	})
}

func (r *employeeRepository) DeleteWithUser(id uuid.UUID) error {
	var employee models.Employee
	if err := r.db.First(&employee, "id = ?", id).Error; err != nil {
		return err
	}
	
	originalEmail := employee.Email

	return r.db.Transaction(func(tx *gorm.DB) error {
		deletedSuffix := fmt.Sprintf("@deleted_%d", time.Now().Unix())

		// Update unique fields to avoid constraint violation on recreate
		employee.Email = employee.Email + deletedSuffix
		employee.EmployeeCode = employee.EmployeeCode + deletedSuffix
		if employee.Phone != "" {
			employee.Phone = employee.Phone + deletedSuffix
		}
		
		if err := tx.Save(&employee).Error; err != nil {
			return fmt.Errorf("failed to update employee unique fields: %w", err)
		}

		if err := tx.Delete(&employee).Error; err != nil {
			return fmt.Errorf("failed to delete employee: %w", err)
		}
		
		// Also find and delete the login user record
		var user models.User
		if err := tx.Where("email = ? AND user_type = ?", originalEmail, 2).First(&user).Error; err == nil {
			user.Email = user.Email + deletedSuffix
			if err := tx.Save(&user).Error; err != nil {
				return fmt.Errorf("failed to update user unique fields: %w", err)
			}
			if err := tx.Delete(&user).Error; err != nil {
				return fmt.Errorf("failed to delete user record: %w", err)
			}
		}
		return nil
	})
}
