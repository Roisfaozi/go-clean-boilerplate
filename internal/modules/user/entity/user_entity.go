package entity

type User struct {
	ID        string `gorm:"column:id;primaryKey"`
	Password  string `gorm:"column:password"`
	Name      string `gorm:"column:name"`
	Token     string `gorm:"column:token"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}
