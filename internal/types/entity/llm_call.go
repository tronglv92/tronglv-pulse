package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type LLMCall struct {
	Base
	TenantID         int64           `gorm:"column:tenant_id;not null;index"`
	Model            string          `gorm:"column:model;not null;size:100"`
	Purpose          string          `gorm:"column:purpose;not null;size:100"`
	PromptTokens     int             `gorm:"column:prompt_tokens;not null;default:0"`
	CompletionTokens int             `gorm:"column:completion_tokens;not null;default:0"`
	TotalTokens      int             `gorm:"column:total_tokens;not null;default:0"`
	CostUSD          decimal.Decimal `gorm:"column:cost_usd;type:numeric(10,6);not null;default:0"`
}

func (LLMCall) TableName() string { return "llm_calls" }

// RCACache is the L2 tier of the three-tier RCA cache (L1=Redis, L2=DB, L3=Claude).
type RCACache struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:id"`
	Fingerprint string    `gorm:"column:fingerprint;not null;size:64;uniqueIndex"`
	Summary     string    `gorm:"column:summary;not null;type:text"`
	Model       string    `gorm:"column:model;not null;size:100"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:now()"`
}

func (RCACache) TableName() string { return "rca_cache" }
