package entity

type EmailVerificationToken struct {
	Email     string `gorm:"primaryKey;column:email"`
	Token     string `gorm:"column:token;index"`
	ExpiresAt int64  `gorm:"column:expires_at"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:milli"`
}

func (EmailVerificationToken) TableName() string {
	return "email_verification_tokens"
}
