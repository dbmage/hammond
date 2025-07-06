package db

import (
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

// Base is
type Base struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index" json:"deletedAt"`
}

// BeforeCreate
func (base *Base) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("id", uuid.New())
	return nil
}
