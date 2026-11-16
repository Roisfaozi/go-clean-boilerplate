package entity

type EmailVerificationToken struct {
	Email     string `db:"email"`
	Token     string `db:"token"`
	ExpiresAt int64  `db:"expires_at"`
	CreatedAt int64  `db:"created_at"`
}

func (EmailVerificationToken) TableName() string {
	return "email_verification_tokens"
}
