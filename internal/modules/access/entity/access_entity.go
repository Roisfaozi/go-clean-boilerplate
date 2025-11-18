package entity

// AccessRight represents an access right entity in the database
type AccessRight struct {
	ID          uint       `gorm:"primaryKey;column:id"`
	Name        string     `gorm:"column:name;type:varchar(191);unique;not null"`
	Description string     `gorm:"column:description;type:text"`
	Endpoints   []Endpoint `gorm:"many2many:access_right_endpoints;"`
	CreatedAt   int64      `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt   int64      `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

func (AccessRight) TableName() string {
	return "access_rights"
}

// Endpoint represents an API endpoint entity in the database
type Endpoint struct {
	ID        uint   `gorm:"primaryKey;column:id"`
	Path      string `gorm:"column:path;type:varchar(191);not null"`
	Method    string `gorm:"column:method;type:varchar(10);not null"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

func (Endpoint) TableName() string {
	return "endpoints"
}

// AccessRightEndpoint represents the junction table for many-to-many relationship
type AccessRightEndpoint struct {
	AccessRightID uint `gorm:"primaryKey;column:access_right_id"`
	EndpointID    uint `gorm:"primaryKey;column:endpoint_id"`
}

func (AccessRightEndpoint) TableName() string {
	return "access_right_endpoints"
}
