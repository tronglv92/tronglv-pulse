package entity

import "github.com/pgvector/pgvector-go"

type ErrorEmbedding struct {
	Base
	TenantID    int64           `gorm:"column:tenant_id;not null;index"`
	Fingerprint string          `gorm:"column:fingerprint;not null;size:64;index"`
	Embedding   pgvector.Vector `gorm:"column:embedding;type:vector(1536);not null"`
	Model       string          `gorm:"column:model;not null;size:100"`
}

func (ErrorEmbedding) TableName() string { return "error_embeddings" }
