package message_model

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestMessageJIDColumnNames(t *testing.T) {
	parsed, err := schema.Parse(&Message{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse message schema: %v", err)
	}

	for fieldName, expected := range map[string]string{
		"ChatJID":        "chat_jid",
		"SenderJID":      "sender_jid",
		"ParticipantJID": "participant_jid",
	} {
		field := parsed.LookUpField(fieldName)
		if field == nil {
			t.Fatalf("field %s not found", fieldName)
		}
		if field.DBName != expected {
			t.Fatalf("field %s mapped to %s, want %s", fieldName, field.DBName, expected)
		}
	}
}
