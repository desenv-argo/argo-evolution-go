package analytics_settings

import (
	"sync/atomic"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const singletonID uint = 1

type CaptureSettings struct {
	ID                    uint      `json:"-" gorm:"primaryKey"`
	MessageCaptureEnabled bool      `json:"message_capture_enabled" gorm:"not null"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (CaptureSettings) TableName() string { return "analytics_settings" }

// CaptureGate keeps the hot event path lock-free while persisting the operator
// choice in PostgreSQL. The environment variable is used only as the first-run
// default; subsequent changes are made from the Manager and survive restarts.
type CaptureGate struct {
	db      *gorm.DB
	enabled atomic.Bool
}

func NewCaptureGate(db *gorm.DB, defaultEnabled bool) (*CaptureGate, error) {
	if err := db.AutoMigrate(&CaptureSettings{}); err != nil {
		return nil, err
	}

	seed := CaptureSettings{ID: singletonID, MessageCaptureEnabled: defaultEnabled}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&seed).Error; err != nil {
		return nil, err
	}

	var settings CaptureSettings
	if err := db.First(&settings, singletonID).Error; err != nil {
		return nil, err
	}

	gate := &CaptureGate{db: db}
	gate.enabled.Store(settings.MessageCaptureEnabled)
	return gate, nil
}

func (g *CaptureGate) Enabled() bool {
	return g != nil && g.enabled.Load()
}

func (g *CaptureGate) Settings() CaptureSettings {
	return CaptureSettings{ID: singletonID, MessageCaptureEnabled: g.Enabled()}
}

func (g *CaptureGate) Update(enabled bool) (*CaptureSettings, error) {
	settings := CaptureSettings{ID: singletonID, MessageCaptureEnabled: enabled}
	result := g.db.Model(&CaptureSettings{}).
		Where("id = ?", singletonID).
		Updates(map[string]interface{}{"message_capture_enabled": enabled})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		if err := g.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"message_capture_enabled"}),
		}).Create(&settings).Error; err != nil {
			return nil, err
		}
	}
	g.enabled.Store(enabled)
	return &settings, nil
}
