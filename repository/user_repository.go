package repository

import (
	"fmt"
	"go-api/models"
	"time"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(id int) (*models.User, error)
	Update(user *models.User) error
	Delete(id int) (int64, error)
	UpdateStatus(id int, isBanned bool) (int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(id int) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id int) (int64, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return 0, err
	}

	deletedSuffix := fmt.Sprintf("@deleted_%d", time.Now().Unix())
	user.Email = user.Email + deletedSuffix

	if err := r.db.Save(&user).Error; err != nil {
		return 0, err
	}

	result := r.db.Delete(&user)
	return result.RowsAffected, result.Error
}

func (r *userRepository) UpdateStatus(id int, isBanned bool) (int64, error) {
	result := r.db.Model(&models.User{}).Where("id = ?", id).Update("is_banned", isBanned)
	return result.RowsAffected, result.Error
}
