package repository

import (
	"go-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskAssignmentRepository interface {
	Create(assignment *models.TaskAssignment) error
	FindByTaskID(taskID uuid.UUID, limit, offset int) ([]models.TaskAssignment, int64, error)
	Delete(id uuid.UUID) error
}

type taskAssignmentRepository struct {
	db *gorm.DB
}

func NewTaskAssignmentRepository(db *gorm.DB) TaskAssignmentRepository {
	return &taskAssignmentRepository{db: db}
}

func (r *taskAssignmentRepository) Create(assignment *models.TaskAssignment) error {
	return r.db.Create(assignment).Error
}

func (r *taskAssignmentRepository) FindByTaskID(taskID uuid.UUID, limit, offset int) ([]models.TaskAssignment, int64, error) {
	var assignments []models.TaskAssignment
	var totalCount int64

	err := r.db.Model(&models.TaskAssignment{}).Where("task_id = ?", taskID).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("task_id = ?", taskID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&assignments).Error
	return assignments, totalCount, err
}

func (r *taskAssignmentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.TaskAssignment{}, "id = ?", id).Error
}
