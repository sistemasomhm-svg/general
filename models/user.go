package models

import (
	"time"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"uniqueIndex;not null"`
	// El Hash derivado que el cliente envía para autenticarse
	AuthHash     string    `gorm:"not null"` 
	// Salt necesario para que el cliente derive su Master Key (se entrega al login)
	ClientSalt   string    `gorm:"not null"` 
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
