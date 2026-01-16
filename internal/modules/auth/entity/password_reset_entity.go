package entity

import "time"

type PasswordResetToken struct {
	Email     string    `gorm:"primaryKey;column:email"`
	Token     string    `gorm:"column:token;index"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}
