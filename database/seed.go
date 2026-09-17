package database

import (
	"fmt"
	"os"
	"strings"

	"go-api/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUser(db *gorm.DB) error {
	email := strings.ToLower(strings.TrimSpace(os.Getenv("SEED_USER_EMAIL")))
	password := os.Getenv("SEED_USER_PASSWORD")
	if email == "" || password == "" {
		return nil
	}

	var user models.User
	err := db.Where("email = ?", email).First(&user).Error
	if err == nil {
		if user.UserType != 0 {
			return db.Model(&user).Update("user_type", 0).Error
		}
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash seed user password: %w", err)
	}

	name := os.Getenv("SEED_USER_NAME")
	if name == "" {
		name = "Shivanshu"
	}

	return db.Create(&models.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		UserType: 0,
	}).Error
}
