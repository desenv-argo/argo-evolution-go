package analytics_model

import (
	"time"

	message_model "github.com/evolution-foundation/evolution-go/pkg/message/model"
)

type Filters struct {
	From        time.Time
	To          time.Time
	InstanceID  string
	Search      string
	Direction   string
	MessageType string
	Status      string
	IsGroup     *bool
	Before      *time.Time
	Limit       int
}

type VolumePoint struct {
	Bucket   time.Time `json:"bucket"`
	Inbound  int64     `json:"inbound"`
	Outbound int64     `json:"outbound"`
	Total    int64     `json:"total"`
}

type DashboardSummary struct {
	From                  time.Time     `json:"from"`
	To                    time.Time     `json:"to"`
	MessageCaptureEnabled bool          `json:"message_capture_enabled"`
	InstancesTotal        int64         `json:"instances_total"`
	InstancesConnected    int64         `json:"instances_connected"`
	InstancesOffline      int64         `json:"instances_offline"`
	ConnectionRate        float64       `json:"connection_rate"`
	MessagesTotal         int64         `json:"messages_total"`
	InboundMessages       int64         `json:"inbound_messages"`
	OutboundMessages      int64         `json:"outbound_messages"`
	ActiveConversations   int64         `json:"active_conversations"`
	UniqueContacts        int64         `json:"unique_contacts"`
	DeliveredMessages     int64         `json:"delivered_messages"`
	ReadMessages          int64         `json:"read_messages"`
	DeliveryRate          float64       `json:"delivery_rate"`
	ReadRate              float64       `json:"read_rate"`
	Volume                []VolumePoint `json:"volume"`
}

type ConversationSummary struct {
	InstanceID        string    `json:"instance_id"`
	InstanceName      string    `json:"instance_name"`
	ChatJID           string    `json:"chat_jid"`
	Contact           string    `json:"contact"`
	PushName          string    `json:"push_name,omitempty"`
	IsGroup           bool      `json:"is_group"`
	LastMessageID     string    `json:"last_message_id"`
	LastMessageType   string    `json:"last_message_type"`
	LastMessageText   string    `json:"last_message_text"`
	LastMessageAt     time.Time `json:"last_message_at"`
	LastDirection     string    `json:"last_direction"`
	LastStatus        string    `json:"last_status"`
	MessageCount      int64     `json:"message_count"`
	InboundCount      int64     `json:"inbound_count"`
	OutboundCount     int64     `json:"outbound_count"`
	UnansweredInbound int64     `json:"unanswered_inbound"`
}

type ConversationPage struct {
	Items      []ConversationSummary `json:"items"`
	HasMore    bool                  `json:"has_more"`
	NextCursor string                `json:"next_cursor,omitempty"`
}

type MessagePage struct {
	Items      []message_model.Message `json:"items"`
	HasMore    bool                    `json:"has_more"`
	NextCursor string                  `json:"next_cursor,omitempty"`
}
