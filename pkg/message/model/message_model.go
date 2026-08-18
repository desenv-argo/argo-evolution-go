package message_model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Message struct {
	Id              string          `json:"id" gorm:"type:uuid;primaryKey"`
	InstanceID      string          `json:"instance_id" gorm:"type:uuid;index:idx_messages_instance_sent_at,priority:1;index:idx_messages_chat_lookup,priority:1;uniqueIndex:idx_messages_instance_message,priority:1"`
	MessageID       string          `json:"message_id" gorm:"unique;uniqueIndex:idx_messages_instance_message,priority:2"`
	ChatJID         string          `json:"chat_jid" gorm:"column:chat_jid;index:idx_messages_chat_lookup,priority:2;size:191"`
	SenderJID       string          `json:"sender_jid,omitempty" gorm:"column:sender_jid;size:191"`
	ParticipantJID  string          `json:"participant_jid,omitempty" gorm:"column:participant_jid;size:191"`
	PushName        string          `json:"push_name,omitempty"`
	Direction       string          `json:"direction" gorm:"size:16"`
	FromMe          bool            `json:"from_me"`
	IsGroup         bool            `json:"is_group"`
	MessageType     string          `json:"message_type" gorm:"size:64"`
	Text            string          `json:"text,omitempty" gorm:"type:text"`
	Caption         string          `json:"caption,omitempty" gorm:"type:text"`
	MediaURL        string          `json:"media_url,omitempty" gorm:"type:text"`
	MimeType        string          `json:"mime_type,omitempty"`
	FileName        string          `json:"file_name,omitempty"`
	FileSize        int64           `json:"file_size,omitempty"`
	QuotedMessageID string          `json:"quoted_message_id,omitempty"`
	SentAt          time.Time       `json:"sent_at" gorm:"index:idx_messages_instance_sent_at,priority:2"`
	DeliveredAt     *time.Time      `json:"delivered_at,omitempty"`
	ReadAt          *time.Time      `json:"read_at,omitempty"`
	Timestamp       string          `json:"timestamp"` // Legacy API compatibility. New queries use SentAt.
	Status          string          `json:"status" gorm:"size:32"`
	Source          string          `json:"source"` // Legacy contact/source field.
	Referral        json.RawMessage `json:"referral,omitempty" gorm:"type:jsonb"`
	Metadata        json.RawMessage `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) (err error) {
	if m.Id == "" {
		m.Id = uuid.New().String()
	}
	return
}
