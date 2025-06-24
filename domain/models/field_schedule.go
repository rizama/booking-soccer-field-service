package models

import (
	"field-service/constants"
	"time"

	"github.com/google/uuid"
)

type FieldSchedule struct {
	ID        uint                          `gorm:"primaryKey;autoIncrement"`
	UUID      uuid.UUID                     `gorm:"type:uuid;not null"`
	FieldID   uint                          `gorm:"type:int;not null"`
	TimeID    uint                          `gorm:"type:int; not null"`
	Date      time.Time                     `gorm:"type:date; not null"`
	Status    constants.FieldScheduleStatus `gorm:"type:int; not null"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time

	// Relation to field table
	Field Field `gorm:"foreignKey:id;references:field_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	// Relation to time table
	Time Time `gorm:"foreignKey:id;references:time_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
