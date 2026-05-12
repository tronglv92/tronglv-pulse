package entity

type User struct {
	Base
	TenantID     int64  `gorm:"column:tenant_id;not null;index"`
	Email        string `gorm:"column:email;not null;uniqueIndex;size:255"`
	PasswordHash string `gorm:"column:password_hash;not null;size:255"`
	Name         string `gorm:"column:name;size:255"`
	IsActive     bool   `gorm:"column:is_active;not null;default:true"`
}

func (User) TableName() string { return "users" }
