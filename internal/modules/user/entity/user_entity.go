package entity

type User struct {
	ID        string `gorm:"column:id;primaryKey"`
	Password  string `gorm:"column:password"`
	Email     string `gorm:"column:email;unique;not null"`
	Username  string `gorm:"column:username;unique;not null"`
	Name      string `gorm:"column:name"`
	Token     string `gorm:"column:token"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt int64  `gorm:"column:deleted_at;autoDeleteTime:milli"`
}
