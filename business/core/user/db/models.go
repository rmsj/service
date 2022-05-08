package db

import (
	"time"

	"github.com/lib/pq"
)

// User represent the structure we need for moving data
// between the app and the database.
type User struct {
	UserID       string         `gorm:"column:user_id;primaryKey"`
	Name         string         `gorm:"column:name"`
	Email        string         `gorm:"column:email"`
	Roles        pq.StringArray `gorm:"column:roles;type:string[]"`
	PasswordHash []byte         `gorm:"column:password_hash"`
	DateCreated  time.Time      `gorm:"column:date_created"`
	DateUpdated  time.Time      `gorm:"column:date_updated"`
}

func (User) TableName() string {
	return "users"
}
