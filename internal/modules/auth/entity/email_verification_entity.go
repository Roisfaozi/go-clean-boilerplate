package entity

type EmailVerificationToken struct {
	Email     string `db:"email" gorm:"primaryKey;column:email"`
	Token     string `db:"token" gorm:"column:token;index"`
	ExpiresAt int64  `db:"expires_at" gorm:"column:expires_at"`
	CreatedAt int64  `db:"created_at" gorm:"column:created_at;autoCreateTime:milli"`
}

func (EmailVerificationToken) TableName() string {
	return "email_verification_tokens"
}
