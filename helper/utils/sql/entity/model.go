package entity

import (
	"pulse/helper/utils/identity"
	"pulse/helper/utils/toolkit/timex"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
	"time"
)

type IdModel struct {
	Id           int32          `gorm:"primary_key:auto_increment"`
	UId          string         `gorm:"uniqueIndex;not null;type:varchar(21)"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at"`
	CreatedByUId string         `gorm:"column:created_by_uid;type:varchar(36);index:idx_created_by_uid;"`
	CreatedBy    string         `gorm:"type:varchar(255)"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedByUId string         `gorm:"column:updated_by_uid;type:varchar(36)"`
	UpdatedBy    string         `gorm:"column:updated_by;type:varchar(255)"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
}

func (m *IdModel) GetId() int32 {
	return m.Id
}

func (m *IdModel) GetUId() string {
	return m.UId
}

func (m *IdModel) GetDeletedAt() gorm.DeletedAt {
	return m.DeletedAt
}

func (m *IdModel) GetCreatedByUId() string {
	return m.CreatedByUId
}

func (m *IdModel) GetCreatedBy() string {
	return m.CreatedBy
}

func (m *IdModel) GetCreatedAt() time.Time {
	return m.CreatedAt
}

func (m *IdModel) GetUpdatedByUId() string {
	return m.UpdatedByUId
}

func (m *IdModel) GetUpdatedBy() string {
	return m.UpdatedBy
}

func (m *IdModel) GetUpdatedAt() time.Time {
	return m.UpdatedAt
}

func (m *IdModel) BeforeCreate(tx *gorm.DB) error {
	if subject, err := identity.MustFromContext(tx.Statement.Context); err == nil {
		tx.Statement.SetColumn("CreatedBy", subject.GetName())
		tx.Statement.SetColumn("CreatedByUId", subject.GetId())
		tx.Statement.SetColumn("UpdatedBy", subject.GetName())
		tx.Statement.SetColumn("UpdatedByUId", subject.GetId())
	}

	if m.UId == "" {
		uid, _ := gonanoid.New()
		tx.Statement.SetColumn("UId", uid)
	}
	if m.CreatedAt.IsZero() {
		tx.Statement.SetColumn("CreatedAt", timex.Now())
	}
	if m.UpdatedAt.IsZero() {
		tx.Statement.SetColumn("UpdatedAt", timex.Now())
	}
	return nil
}

func (m *IdModel) BeforeUpdate(tx *gorm.DB) error {
	if subject, err := identity.MustFromContext(tx.Statement.Context); err == nil {
		tx.Statement.SetColumn("UpdatedBy", subject.GetName())
		tx.Statement.SetColumn("UpdatedByUId", subject.GetId())
	}
	if m.UpdatedAt.IsZero() {
		tx.Statement.SetColumn("UpdatedAt", timex.Now())
	}
	return nil
}
