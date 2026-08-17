package message_normalizer

import (
	"testing"
	"time"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func TestNormalizeTextMessage(t *testing.T) {
	sentAt := time.Date(2026, time.August, 17, 12, 30, 0, 0, time.FixedZone("BRT", -3*60*60))
	info := types.MessageInfo{
		MessageSource: types.MessageSource{
			Chat:     types.JID{User: "5531999999999", Server: types.DefaultUserServer},
			Sender:   types.JID{User: "5531999999999", Server: types.DefaultUserServer},
			IsFromMe: false,
			IsGroup:  false,
		},
		ID:        "message-1",
		Timestamp: sentAt,
		PushName:  "Cliente Argo",
	}
	waMessage := &waE2E.Message{ExtendedTextMessage: &waE2E.ExtendedTextMessage{
		Text: proto.String("Preciso de ajuda"),
		ContextInfo: &waE2E.ContextInfo{
			StanzaID: proto.String("quoted-1"),
		},
	}}

	message := Normalize("d7da53f2-3211-4698-a0a9-ffdccfcb62fc", info, waMessage, nil)

	if message.Direction != "inbound" || message.Status != "Received" {
		t.Fatalf("unexpected direction/status: %s/%s", message.Direction, message.Status)
	}
	if message.Text != "Preciso de ajuda" || message.MessageType != "text" {
		t.Fatalf("unexpected normalized content: %#v", message)
	}
	if message.QuotedMessageID != "quoted-1" {
		t.Fatalf("expected quoted message id, got %q", message.QuotedMessageID)
	}
	if !message.SentAt.Equal(sentAt.UTC()) {
		t.Fatalf("expected UTC sent time, got %s", message.SentAt)
	}
}

func TestApplyMediaMetadata(t *testing.T) {
	message := Normalize("d7da53f2-3211-4698-a0a9-ffdccfcb62fc", types.MessageInfo{
		MessageSource: types.MessageSource{
			Chat:   types.JID{User: "5531999999999", Server: types.DefaultUserServer},
			Sender: types.JID{User: "5531999999999", Server: types.DefaultUserServer},
		},
		ID:        "image-1",
		Timestamp: time.Now(),
	}, &waE2E.Message{ImageMessage: &waE2E.ImageMessage{
		Caption:  proto.String("Comprovante"),
		Mimetype: proto.String("image/jpeg"),
	}}, nil)

	ApplyMediaMetadata(&message, map[string]interface{}{
		"mediaUrl": "https://storage.example/image-1.jpg",
		"mimetype": "image/jpeg",
	})

	if message.MediaURL == "" || message.Caption != "Comprovante" || message.MessageType != "image" {
		t.Fatalf("unexpected media normalization: %#v", message)
	}
}
