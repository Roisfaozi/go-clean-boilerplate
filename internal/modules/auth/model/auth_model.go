package model

import (
	"time"
)

type LoginRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=50,xss"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
	IPAddress string `json:"-"`
	UserAgent string `json:"-"`
}

type RegisterRequest struct {
	Name      string `json:"name" validate:"required,min=3,max=100,xss"`
	Username  string `json:"username" validate:"required,min=3,max=50,xss"`
	Email     string `json:"email" validate:"required,email,max=100"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
	IPAddress string `json:"-"`
	UserAgent string `json:"-"`
}

type Auth struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	SessionID    string    `json:"session_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

type UserInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	AvatarURL string `json:"avatar_url"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,max=500"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email,max=100"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required,max=500"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required,max=500"`
}

type UserSessionContext struct {
	UserID    string
	OrgID     string
	SessionID string
	Role      string
	Username  string
}
