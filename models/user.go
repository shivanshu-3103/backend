package models

import "time"

type User struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	UserType  int       `json:"user_type" gorm:"not null;default:1"`
	Password  string    `json:"password" gorm:"not null"`
	IsBanned  bool      `json:"is_banned" gorm:"not null;default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}