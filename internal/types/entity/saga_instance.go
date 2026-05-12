package entity

import (
	"encoding/json"

	"pulse/internal/types/define/enum"
)

type SagaInstance struct {
	Base
	SagaType    string          `gorm:"column:saga_type;not null;size:100;index"`
	Status      enum.SagaStatus `gorm:"column:status;not null;default:0;index"`
	Payload     json.RawMessage `gorm:"column:payload;type:jsonb;not null"`
	CurrentStep string          `gorm:"column:current_step;size:100"`
	Error       string          `gorm:"column:error;type:text"`
}

func (SagaInstance) TableName() string { return "saga_instances" }
