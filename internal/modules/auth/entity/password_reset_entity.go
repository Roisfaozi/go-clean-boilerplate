package entity

type PasswordResetToken struct {
	Email     string `db:"email"`
	Token     string `db:"token"`
	ExpiresAt int64  `db:"expires_at"`
	CreatedAt int64  `db:"created_at"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}
