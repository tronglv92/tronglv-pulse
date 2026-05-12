package entity

import (
	"time"

	"pulse/internal/types/define/enum"
)

type Incident struct {
	Base
	TenantID    int64              `gorm:"column:tenant_id;not null;index"`
	Title       string             `gorm:"column:title;not null;type:text"`
	Description string             `gorm:"column:description;type:text"`
	ServiceName string             `gorm:"column:service_name;not null;size:255;index"`
	Fingerprint string             `gorm:"column:fingerprint;not null;size:64;index"`
	Status      enum.IncidentStatus `gorm:"column:status;not null;default:'open';type:incident_status;index"`
	Severity    int                `gorm:"column:severity;not null;default:1"`
	RCASummary  string             `gorm:"column:rca_summary;type:text"`
	ResolvedAt  *time.Time         `gorm:"column:resolved_at"`
}

func (Incident) TableName() string { return "incidents" }
