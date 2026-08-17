package message_repository

import (
	message_model "github.com/evolution-foundation/evolution-go/pkg/message/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MessageRepository interface {
	InsertMessage(message message_model.Message) error
	GetMessageByID(messageID string) (*message_model.Message, error)
	DeleteAllMessages() (int64, error)
	GetLatestMessageID(source string) (string, string, error)
}

type messageRepository struct {
	db *gorm.DB
}

func messageUpdateColumns(message message_model.Message) []string {
	updates := []string{"timestamp"}
	if message.Status != "" {
		updates = append(updates, "status")
	}
	if message.Source != "" {
		updates = append(updates, "source")
	}
	if message.InstanceID != "" {
		updates = append(updates, "instance_id")
	}
	if message.ChatJID != "" {
		updates = append(updates, "chat_jid")
	}
	if message.SenderJID != "" {
		updates = append(updates, "sender_jid")
	}
	if message.ParticipantJID != "" {
		updates = append(updates, "participant_jid")
	}
	if message.PushName != "" {
		updates = append(updates, "push_name")
	}
	if message.Direction != "" {
		updates = append(updates, "direction", "from_me", "is_group")
	}
	if message.MessageType != "" {
		updates = append(updates, "message_type")
	}
	if message.Text != "" {
		updates = append(updates, "text")
	}
	if message.Caption != "" {
		updates = append(updates, "caption")
	}
	if message.MediaURL != "" {
		updates = append(updates, "media_url")
	}
	if message.MimeType != "" {
		updates = append(updates, "mime_type")
	}
	if message.FileName != "" {
		updates = append(updates, "file_name")
	}
	if message.FileSize > 0 {
		updates = append(updates, "file_size")
	}
	if message.QuotedMessageID != "" {
		updates = append(updates, "quoted_message_id")
	}
	if !message.SentAt.IsZero() {
		updates = append(updates, "sent_at")
	}
	if message.DeliveredAt != nil {
		updates = append(updates, "delivered_at")
	}
	if message.ReadAt != nil {
		updates = append(updates, "read_at")
	}
	if len(message.Referral) > 0 {
		updates = append(updates, "referral")
	}
	if len(message.Metadata) > 0 {
		updates = append(updates, "metadata")
	}

	return updates
}

func (m *messageRepository) InsertMessage(message message_model.Message) error {
	assignments := make(map[string]interface{})
	for _, column := range messageUpdateColumns(message) {
		if column == "status" {
			assignments[column] = gorm.Expr(`CASE
				WHEN messages.status = 'Read' THEN messages.status
				WHEN messages.status = 'Delivered' AND EXCLUDED.status IN ('Sent', 'Received') THEN messages.status
				ELSE EXCLUDED.status
			END`)
			continue
		}
		assignments[column] = gorm.Expr("EXCLUDED." + column)
	}

	return m.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "message_id"}},
		DoUpdates: clause.Assignments(assignments),
	}).Create(&message).Error
}

func (m *messageRepository) GetMessageByID(messageID string) (*message_model.Message, error) {
	var message message_model.Message
	err := m.db.Where("message_id = ?", messageID).First(&message).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &message, nil
}

func (m *messageRepository) DeleteAllMessages() (int64, error) {
	result := m.db.Exec("DELETE FROM messages")
	return result.RowsAffected, result.Error
}

func (m *messageRepository) GetLatestMessageID(source string) (string, string, error) {
	var message message_model.Message
	err := m.db.Where("source = ?", source).Order("timestamp DESC").First(&message).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", "", nil
		}
		return "", "", err
	}

	return message.MessageID, message.Timestamp, nil
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}
