package database

import (
	"fmt"
	"os"
	"time"

	"go-api/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectPostgres() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")

		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host,
			port,
			user,
			password,
			dbname,
			sslmode,
		)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// Auto Migrate the models
	if err := db.AutoMigrate(&models.User{}, &models.Employee{}, &models.Department{}, &models.Role{}, &models.EmployeeDevice{}, &models.EmployeeScreenshot{}, &models.EmployeeActivity{}, &models.Application{}, &models.Website{}, &models.Team{}, &models.Task{}, &models.TaskAssignment{}, &models.Attendance{}, &models.AttendanceEvent{}); err != nil {
		return nil, err
	}
	// image_data was used by the old local-storage implementation and is no longer populated.
	if err := db.Exec("ALTER TABLE employee_screenshots DROP COLUMN IF EXISTS image_data").Error; err != nil {
		return nil, err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_employee_devices_employee_id ON employee_devices (employee_id)").Error; err != nil {
		return nil, err
	}
	return db, nil
}
