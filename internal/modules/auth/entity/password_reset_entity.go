package entity

type PasswordResetToken struct {
	Email     string `gorm:"primaryKey;column:email"`
	Token     string `gorm:"column:token;index"`
	ExpiresAt int64  `gorm:"column:expires_at"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:milli"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}
