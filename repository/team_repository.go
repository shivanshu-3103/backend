package repository

import (
	"go-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamRepository interface {
	Create(team *models.Team) error
	FindByID(id uuid.UUID) (*models.Team, error)
	FindAll(limit, offset int) ([]models.Team, int64, error)
	Update(team *models.Team) error
	Delete(id uuid.UUID) error
}

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) Create(team *models.Team) error {
	return r.db.Create(team).Error
}

func (r *teamRepository) FindByID(id uuid.UUID) (*models.Team, error) {
	var team models.Team
	err := r.db.First(&team, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) FindAll(limit, offset int) ([]models.Team, int64, error) {
	var teams []models.Team
	var totalCount int64

	err := r.db.Model(&models.Team{}).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Order("name ASC").Limit(limit).Offset(offset).Find(&teams).Error
	return teams, totalCount, err
}

func (r *teamRepository) Update(team *models.Team) error {
	return r.db.Save(team).Error
}

func (r *teamRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Team{}, "id = ?", id).Error
}
