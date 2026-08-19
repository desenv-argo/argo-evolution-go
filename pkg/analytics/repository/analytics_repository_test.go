package analytics_repository

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestConversationRowMapsChatJIDColumn(t *testing.T) {
	parsed, err := schema.Parse(&conversationRow{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse conversation row schema: %v", err)
	}

	field := parsed.LookUpField("ChatJID")
	if field == nil {
		t.Fatal("ChatJID field was not found")
	}
	if field.DBName != "chat_jid" {
		t.Fatalf("ChatJID mapped to %q, want chat_jid", field.DBName)
	}
}

func TestConversationContactDoesNotPresentLIDAsPhone(t *testing.T) {
	if got := conversationContact("189386650058872@lid", "189386650058872"); got != "" {
		t.Fatalf("conversationContact() = %q, want empty contact for unresolved LID", got)
	}
}

func TestConversationContactPreservesPhoneJIDSource(t *testing.T) {
	if got := conversationContact("5582999999999@s.whatsapp.net", "5582999999999"); got != "5582999999999" {
		t.Fatalf("conversationContact() = %q, want source phone", got)
	}
}
