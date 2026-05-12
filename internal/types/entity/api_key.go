package entity

import "time"

type APIKey struct {
	Base
	TenantID   int64      `gorm:"column:tenant_id;not null;index"`
	KeyHash    string     `gorm:"column:key_hash;not null;size:255;uniqueIndex:idx_api_keys_hash,where:deleted_at IS NULL"`
	Name       string     `gorm:"column:name;size:100"`
	LastUsedAt *time.Time `gorm:"column:last_used_at"`
	ExpiresAt  *time.Time `gorm:"column:expires_at"`
}

func (APIKey) TableName() string { return "api_keys" }
