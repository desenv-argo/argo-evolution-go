package message_normalizer

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	message_model "github.com/evolution-foundation/evolution-go/pkg/message/model"
	"github.com/evolution-foundation/evolution-go/pkg/utils"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// Normalize converts a whatsmeow event into the stable, query-friendly shape
// used by the Manager. The raw protobuf remains on the webhook path; it is not
// stored in PostgreSQL because it may contain thumbnails and other large data.
func Normalize(instanceID string, info types.MessageInfo, waMessage *waE2E.Message, referral json.RawMessage) message_model.Message {
	sentAt := info.Timestamp.UTC()
	if sentAt.IsZero() {
		sentAt = time.Now().UTC()
	}

	direction := "inbound"
	status := "Received"
	if info.IsFromMe {
		direction = "outbound"
		status = "Sent"
	}

	messageType := normalizeType(utils.GetMessageType(waMessage))
	text, caption, metadata := extractContent(waMessage)
	mimeType, fileName, fileSize := extractFileMetadata(waMessage)

	participantJID := ""
	if info.IsGroup {
		participantJID = info.Sender.String()
	}

	return message_model.Message{
		InstanceID:      instanceID,
		MessageID:       info.ID,
		ChatJID:         info.Chat.String(),
		SenderJID:       info.Sender.String(),
		ParticipantJID:  participantJID,
		PushName:        info.PushName,
		Direction:       direction,
		FromMe:          info.IsFromMe,
		IsGroup:         info.IsGroup,
		MessageType:     messageType,
		Text:            text,
		Caption:         caption,
		MimeType:        mimeType,
		FileName:        fileName,
		FileSize:        fileSize,
		QuotedMessageID: extractQuotedMessageID(waMessage),
		SentAt:          sentAt,
		Timestamp:       sentAt.Format("2006-01-02 15:04:05"),
		Status:          status,
		Source:          info.Chat.ToNonAD().User,
		Referral:        referral,
		Metadata:        metadata,
	}
}

// ApplyMediaMetadata enriches a normalized message with the storage metadata
// already produced by the existing webhook media pipeline.
func ApplyMediaMetadata(message *message_model.Message, webhookMessage map[string]interface{}) {
	if message == nil || webhookMessage == nil {
		return
	}

	if value, ok := webhookMessage["mediaUrl"].(string); ok {
		message.MediaURL = value
	}
	if value, ok := webhookMessage["mimetype"].(string); ok && value != "" {
		message.MimeType = value
	}
}

func normalizeType(raw string) string {
	switch {
	case raw == "text":
		return "text"
	case strings.HasPrefix(raw, "image "):
		return "image"
	case strings.HasPrefix(raw, "sticker "):
		return "sticker"
	case strings.HasPrefix(raw, "video "), strings.HasPrefix(raw, "round video "):
		return "video"
	case strings.HasPrefix(raw, "audio "):
		return "audio"
	case strings.HasPrefix(raw, "document "):
		return "document"
	case strings.HasPrefix(raw, "contact"):
		return "contact"
	case strings.HasPrefix(raw, "location"), strings.HasPrefix(raw, "live location"):
		return "location"
	case strings.HasPrefix(raw, "poll"):
		return "poll"
	case strings.Contains(raw, "response"), strings.Contains(raw, "reply"):
		return "interaction"
	default:
		return raw
	}
}

func extractContent(message *waE2E.Message) (string, string, json.RawMessage) {
	if message == nil {
		return "", "", nil
	}

	switch {
	case message.Conversation != nil:
		return message.GetConversation(), "", nil
	case message.ExtendedTextMessage != nil:
		return message.GetExtendedTextMessage().GetText(), "", nil
	case message.ImageMessage != nil:
		return "", message.GetImageMessage().GetCaption(), nil
	case message.VideoMessage != nil:
		return "", message.GetVideoMessage().GetCaption(), nil
	case message.PtvMessage != nil:
		return "", message.GetPtvMessage().GetCaption(), nil
	case message.DocumentMessage != nil:
		return "", message.GetDocumentMessage().GetCaption(), nil
	case message.AudioMessage != nil:
		return "", "", nil
	case message.ContactMessage != nil:
		return message.GetContactMessage().GetDisplayName(), "", nil
	case message.LocationMessage != nil:
		location := message.GetLocationMessage()
		return fmt.Sprintf("Localizacao: %.6f, %.6f", location.GetDegreesLatitude(), location.GetDegreesLongitude()), "", nil
	case message.ButtonsResponseMessage != nil:
		response := message.GetButtonsResponseMessage()
		return response.GetSelectedDisplayText(), "", marshalMetadata(map[string]interface{}{
			"button_id": response.GetSelectedButtonID(),
			"kind":      "buttons_response",
		})
	case message.TemplateButtonReplyMessage != nil:
		response := message.GetTemplateButtonReplyMessage()
		return response.GetSelectedDisplayText(), "", marshalMetadata(map[string]interface{}{
			"button_id": response.GetSelectedID(),
			"kind":      "template_button_reply",
		})
	case message.ListResponseMessage != nil:
		response := message.GetListResponseMessage()
		selectedID := ""
		if response.GetSingleSelectReply() != nil {
			selectedID = response.GetSingleSelectReply().GetSelectedRowID()
		}
		return response.GetTitle(), response.GetDescription(), marshalMetadata(map[string]interface{}{
			"button_id": selectedID,
			"kind":      "list_response",
		})
	case message.InteractiveResponseMessage != nil:
		response := message.GetInteractiveResponseMessage()
		if nativeFlow := response.GetNativeFlowResponseMessage(); nativeFlow != nil {
			return nativeFlow.GetName(), "", marshalMetadata(map[string]interface{}{
				"kind":        "native_flow_response",
				"params_json": nativeFlow.GetParamsJSON(),
			})
		}
	case message.PollCreationMessage != nil:
		return message.GetPollCreationMessage().GetName(), "", nil
	case message.ReactionMessage != nil:
		return message.GetReactionMessage().GetText(), "", nil
	}

	return "", "", nil
}

func extractFileMetadata(message *waE2E.Message) (string, string, int64) {
	if message == nil {
		return "", "", 0
	}

	switch {
	case message.ImageMessage != nil:
		item := message.GetImageMessage()
		return item.GetMimetype(), "", int64(item.GetFileLength())
	case message.VideoMessage != nil:
		item := message.GetVideoMessage()
		return item.GetMimetype(), "", int64(item.GetFileLength())
	case message.PtvMessage != nil:
		item := message.GetPtvMessage()
		return item.GetMimetype(), "", int64(item.GetFileLength())
	case message.AudioMessage != nil:
		item := message.GetAudioMessage()
		return item.GetMimetype(), "", int64(item.GetFileLength())
	case message.DocumentMessage != nil:
		item := message.GetDocumentMessage()
		return item.GetMimetype(), item.GetFileName(), int64(item.GetFileLength())
	case message.StickerMessage != nil:
		item := message.GetStickerMessage()
		return item.GetMimetype(), "", int64(item.GetFileLength())
	default:
		return "", "", 0
	}
}

func extractQuotedMessageID(message *waE2E.Message) string {
	contextInfo := contextInfo(message)
	if contextInfo == nil {
		return ""
	}
	return contextInfo.GetStanzaID()
}

func contextInfo(message *waE2E.Message) *waE2E.ContextInfo {
	if message == nil {
		return nil
	}

	switch {
	case message.ExtendedTextMessage != nil:
		return message.GetExtendedTextMessage().GetContextInfo()
	case message.ImageMessage != nil:
		return message.GetImageMessage().GetContextInfo()
	case message.VideoMessage != nil:
		return message.GetVideoMessage().GetContextInfo()
	case message.PtvMessage != nil:
		return message.GetPtvMessage().GetContextInfo()
	case message.AudioMessage != nil:
		return message.GetAudioMessage().GetContextInfo()
	case message.DocumentMessage != nil:
		return message.GetDocumentMessage().GetContextInfo()
	case message.StickerMessage != nil:
		return message.GetStickerMessage().GetContextInfo()
	case message.LocationMessage != nil:
		return message.GetLocationMessage().GetContextInfo()
	case message.ContactMessage != nil:
		return message.GetContactMessage().GetContextInfo()
	case message.ButtonsResponseMessage != nil:
		return message.GetButtonsResponseMessage().GetContextInfo()
	case message.TemplateButtonReplyMessage != nil:
		return message.GetTemplateButtonReplyMessage().GetContextInfo()
	case message.ListResponseMessage != nil:
		return message.GetListResponseMessage().GetContextInfo()
	case message.InteractiveResponseMessage != nil:
		return message.GetInteractiveResponseMessage().GetContextInfo()
	case message.PollCreationMessage != nil:
		return message.GetPollCreationMessage().GetContextInfo()
	default:
		return nil
	}
}

func marshalMetadata(value map[string]interface{}) json.RawMessage {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return encoded
}
