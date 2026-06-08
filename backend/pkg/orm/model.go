package orm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ID contains a UUID primary key for GORM models.
type ID struct {
	// ID is the UUID primary key.
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`
}

// BeforeCreate assigns a UUID primary key when no ID is present.
func (model *ID) BeforeCreate(*gorm.DB) error {
	if model.ID == uuid.Nil {
		model.ID = uuid.New()
	}
	return nil
}

// Timestamps contains standard persistence timestamps.
type Timestamps struct {
	// CreatedAt is the creation timestamp managed by GORM.
	CreatedAt time.Time

	// UpdatedAt is the update timestamp managed by GORM.
	UpdatedAt time.Time
}

// SoftDelete contains the standard GORM soft delete marker.
type SoftDelete struct {
	// DeletedAt is set when a row is soft deleted.
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
