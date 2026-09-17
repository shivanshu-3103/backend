package main

import (
	"github.com/joho/godotenv"
	"go-api/database"
	"go-api/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		utils.LogInfo("Warning: .env file not found")
	}

	db, err := database.ConnectPostgres()
	if err != nil {
		utils.LogFatal("Database connection failed: %v", err)
	}

	err = db.Exec("ALTER TABLE employee_screenshots DROP COLUMN IF EXISTS image_data;").Error
	if err != nil {
		utils.LogFatal("Error dropping column: %v", err)
	}

	utils.LogSuccess("Successfully dropped image_data column")
}
