package model

type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

type VerifyUserRequest struct {
	Token string `validate:"required,max=100"`
}

type RegisterUserRequest struct {
	Username  string `json:"username" validate:"required,min=6,max=100"`
	Password  string `json:"password" validate:"required,min=8,max=100"`
	Name      string `json:"fullname" validate:"required,min=3,max=100"`
	Email     string `json:"email" validate:"email"`
	IPAddress string `json:"-"`
	UserAgent string `json:"-"`
}

type UpdateUserRequest struct {
	ID        string `json:"-" validate:"required,max=100"`
	Username  string `json:"username" validate:"required,min=6,max=100"`
	Password  string `json:"password,omitempty" validate:"max=100"`
	Name      string `json:"name,omitempty" validate:"max=100"`
	IPAddress string `json:"-"`
	UserAgent string `json:"-"`
}

type GetUserRequest struct {
	ID string `json:"id" validate:"required,max=100"`
}

type GetUserListRequest struct {
	Page     int    `form:"page" json:"page"`
	Limit    int    `form:"limit" json:"limit"`
	Username string `form:"username" json:"username"`
	Email    string `form:"email" json:"email"`
}

type DeleteUserRequest struct {
	ID        string `json:"-" validate:"required"`
	IPAddress string `json:"-"`
	UserAgent string `json:"-"`
}

type UpdateUserStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active suspended banned"`
}
