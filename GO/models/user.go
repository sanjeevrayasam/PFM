package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" swaggerignore:"true"`
	Username  string         `gorm:"uniqueIndex" json:"username"`
	Password  string         `json:"-"` // "-" in json tag to exclude from json responses
}
