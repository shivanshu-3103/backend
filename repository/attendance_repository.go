package repository

import (
	"time"

	"go-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendanceRepository interface {
	FindTodayByEmployeeID(employeeID uuid.UUID, today time.Time) (*models.Attendance, error)
	FindAllByEmployeeID(employeeID uuid.UUID, limit, offset int) ([]models.Attendance, int64, error)
	FindAll(limit, offset int) ([]models.Attendance, int64, error)
	Create(attendance *models.Attendance) error
	Update(attendance *models.Attendance) error
	CreateEvent(event *models.AttendanceEvent) error
}

type attendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db: db}
}

func (r *attendanceRepository) FindTodayByEmployeeID(employeeID uuid.UUID, today time.Time) (*models.Attendance, error) {
	var attendance models.Attendance
	err := r.db.Preload("Events").Where("employee_id = ? AND date = ?", employeeID, today.Format("2006-01-02")).First(&attendance).Error
	if err != nil {
		return nil, err
	}
	return &attendance, nil
}

func (r *attendanceRepository) FindAllByEmployeeID(employeeID uuid.UUID, limit, offset int) ([]models.Attendance, int64, error) {
	var attendances []models.Attendance
	var totalCount int64

	if err := r.db.Model(&models.Attendance{}).Where("employee_id = ?", employeeID).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("Events").Preload("Employee").Where("employee_id = ?", employeeID).Order("date DESC").Limit(limit).Offset(offset).Find(&attendances).Error
	return attendances, totalCount, err
}

func (r *attendanceRepository) FindAll(limit, offset int) ([]models.Attendance, int64, error) {
	var attendances []models.Attendance
	var totalCount int64

	if err := r.db.Model(&models.Attendance{}).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("Events").Preload("Employee").Order("date DESC").Limit(limit).Offset(offset).Find(&attendances).Error
	return attendances, totalCount, err
}

func (r *attendanceRepository) Create(attendance *models.Attendance) error {
	return r.db.Create(attendance).Error
}

func (r *attendanceRepository) Update(attendance *models.Attendance) error {
	return r.db.Save(attendance).Error
}

func (r *attendanceRepository) CreateEvent(event *models.AttendanceEvent) error {
	return r.db.Create(event).Error
}
