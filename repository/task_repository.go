package repository

import (
	"go-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(task *models.Task) error
	FindByID(id uuid.UUID) (*models.Task, error)
	FindAll(limit, offset int) ([]models.Task, int64, error)
	Update(task *models.Task) error
	Delete(id uuid.UUID) error
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(task *models.Task) error {
	return r.db.Create(task).Error
}

func (r *taskRepository) FindByID(id uuid.UUID) (*models.Task, error) {
	var task models.Task
	err := r.db.Preload("Creator").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) FindAll(limit, offset int) ([]models.Task, int64, error) {
	var tasks []models.Task
	var totalCount int64

	if err := r.db.Model(&models.Task{}).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("Creator").Limit(limit).Offset(offset).Find(&tasks).Error
	return tasks, totalCount, err
}

func (r *taskRepository) Update(task *models.Task) error {
	return r.db.Save(task).Error
}

func (r *taskRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Task{}, "id = ?", id).Error
}
