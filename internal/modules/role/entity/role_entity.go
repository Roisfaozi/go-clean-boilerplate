package entity

import "time"

// Role represents the roles table in the database.
type Role struct {
	ID          string    `gorm:"type:varchar(36);primary_key"`
	Name        string    `gorm:"type:varchar(50);not null;unique"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	UpdatedAt   time.Time `gorm:"not null;default:now()"`
}