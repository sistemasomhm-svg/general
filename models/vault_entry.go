package models

import (
	"time"
	"github.com/google/uuid"
)

type VaultEntry struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"` 
	UserID        uuid.UUID `gorm:"type:uuid;index;not null"`
	EncryptedData string    `gorm:"type:text;not null"`
	Version       int64     `gorm:"index;not null;default:1"`
	IsDeleted     bool      `gorm:"default:false"`
	UpdatedAt     time.Time `gorm:"index"`
}
