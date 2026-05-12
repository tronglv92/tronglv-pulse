package entity

type Tenant struct {
	Base
	Name      string `gorm:"column:name;not null;size:255"`
	Slug      string `gorm:"column:slug;not null;size:255;uniqueIndex:idx_tenants_slug,where:deleted_at IS NULL"`
	APISecret string `gorm:"column:api_secret;not null;size:255"`
	IsActive  bool   `gorm:"column:is_active;not null;default:true"`
}

func (Tenant) TableName() string { return "tenants" }
