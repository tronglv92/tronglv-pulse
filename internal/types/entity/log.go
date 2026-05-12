package entity

import "encoding/json"

type LogEntry struct {
	HypertableBase
	TenantID    int64           `gorm:"column:tenant_id;not null;index:idx_log_entries_tenant_id"`
	ServiceName string          `gorm:"column:service_name;not null;size:255;index:idx_log_entries_service_name"`
	Level       string          `gorm:"column:level;not null;size:20;index:idx_log_entries_level"`
	Message     string          `gorm:"column:message;not null;type:text"`
	Fingerprint string          `gorm:"column:fingerprint;not null;size:64;index"`
	Metadata    json.RawMessage `gorm:"column:metadata;type:jsonb"`
}

func (LogEntry) TableName() string { return "log_entries" }

type AnomalyEvent struct {
	HypertableBase
	TenantID    int64   `gorm:"column:tenant_id;not null;index"`
	ServiceName string  `gorm:"column:service_name;not null;size:255;index"`
	Fingerprint string  `gorm:"column:fingerprint;not null;size:64;index"`
	ZScore      float64 `gorm:"column:z_score;not null"`
	WindowSize  int     `gorm:"column:window_size;not null"`
	LogCount    int     `gorm:"column:log_count;not null"`
}

func (AnomalyEvent) TableName() string { return "anomaly_events" }
