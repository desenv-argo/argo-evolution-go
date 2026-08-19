package analytics_repository

import (
	"context"
	"strings"
	"time"

	analytics_model "github.com/evolution-foundation/evolution-go/pkg/analytics/model"
	instance_model "github.com/evolution-foundation/evolution-go/pkg/instance/model"
	message_model "github.com/evolution-foundation/evolution-go/pkg/message/model"
	"gorm.io/gorm"
)

type AnalyticsRepository interface {
	Dashboard(ctx context.Context, filters analytics_model.Filters) (*analytics_model.DashboardSummary, error)
	ListConversations(ctx context.Context, filters analytics_model.Filters) (*analytics_model.ConversationPage, error)
	ListMessages(ctx context.Context, filters analytics_model.Filters, chatJID string) (*analytics_model.MessagePage, error)
}

type analyticsRepository struct {
	db *gorm.DB
}

type conversationRow struct {
	InstanceID        string
	InstanceName      string
	ChatJID           string `gorm:"column:chat_jid"`
	Source            string
	PushName          string
	IsGroup           bool
	MessageID         string
	MessageType       string
	Text              string
	Caption           string
	SentAt            time.Time
	Direction         string
	Status            string
	MessageCount      int64
	InboundCount      int64
	OutboundCount     int64
	UnansweredInbound int64
}

func (r *analyticsRepository) Dashboard(ctx context.Context, filters analytics_model.Filters) (*analytics_model.DashboardSummary, error) {
	summary := &analytics_model.DashboardSummary{From: filters.From, To: filters.To, Volume: []analytics_model.VolumePoint{}}

	instanceQuery := r.db.WithContext(ctx).Model(&instance_model.Instance{})
	if filters.InstanceID != "" {
		instanceQuery = instanceQuery.Where("id = ?", filters.InstanceID)
	}

	type instanceMetrics struct {
		Total     int64
		Connected int64
	}
	var instances instanceMetrics
	if err := instanceQuery.Select(`
		COUNT(*) AS total,
		COALESCE(SUM(CASE WHEN connected = TRUE THEN 1 ELSE 0 END), 0) AS connected
	`).Scan(&instances).Error; err != nil {
		return nil, err
	}
	summary.InstancesTotal = instances.Total
	summary.InstancesConnected = instances.Connected
	summary.InstancesOffline = instances.Total - instances.Connected
	summary.ConnectionRate = percentage(instances.Connected, instances.Total)

	type messageMetrics struct {
		Total               int64
		Inbound             int64
		Outbound            int64
		ActiveConversations int64
		UniqueContacts      int64
		Delivered           int64
		Read                int64
	}
	var messages messageMetrics
	messageQuery := r.scopedMessages(ctx, filters)
	if err := messageQuery.Select(`
		COUNT(*) AS total,
		COALESCE(SUM(CASE WHEN direction = 'inbound' THEN 1 ELSE 0 END), 0) AS inbound,
		COALESCE(SUM(CASE WHEN direction = 'outbound' THEN 1 ELSE 0 END), 0) AS outbound,
		COUNT(DISTINCT CASE WHEN chat_jid <> '' THEN CONCAT(instance_id, ':', chat_jid) END) AS active_conversations,
		COUNT(DISTINCT CASE WHEN is_group = FALSE AND chat_jid <> '' THEN CONCAT(instance_id, ':', chat_jid) END) AS unique_contacts,
		COALESCE(SUM(CASE WHEN direction = 'outbound' AND status IN ('Delivered', 'Read') THEN 1 ELSE 0 END), 0) AS delivered,
		COALESCE(SUM(CASE WHEN direction = 'outbound' AND status = 'Read' THEN 1 ELSE 0 END), 0) AS read
	`).Scan(&messages).Error; err != nil {
		return nil, err
	}

	summary.MessagesTotal = messages.Total
	summary.InboundMessages = messages.Inbound
	summary.OutboundMessages = messages.Outbound
	summary.ActiveConversations = messages.ActiveConversations
	summary.UniqueContacts = messages.UniqueContacts
	summary.DeliveredMessages = messages.Delivered
	summary.ReadMessages = messages.Read
	summary.DeliveryRate = percentage(messages.Delivered, messages.Outbound)
	summary.ReadRate = percentage(messages.Read, messages.Outbound)

	volumeQuery := r.scopedMessages(ctx, filters)
	if err := volumeQuery.Select(`
		DATE_TRUNC('day', sent_at) AS bucket,
		COALESCE(SUM(CASE WHEN direction = 'inbound' THEN 1 ELSE 0 END), 0) AS inbound,
		COALESCE(SUM(CASE WHEN direction = 'outbound' THEN 1 ELSE 0 END), 0) AS outbound,
		COUNT(*) AS total
	`).Group("DATE_TRUNC('day', sent_at)").Order("bucket ASC").Scan(&summary.Volume).Error; err != nil {
		return nil, err
	}

	return summary, nil
}

func (r *analyticsRepository) ListConversations(ctx context.Context, filters analytics_model.Filters) (*analytics_model.ConversationPage, error) {
	limit := normalizeLimit(filters.Limit)
	base := r.scopedMessages(ctx, filters)
	base = applyContentFilters(base, filters)

	ranked := base.Select(`
		messages.*,
		ROW_NUMBER() OVER (PARTITION BY instance_id, chat_jid ORDER BY sent_at DESC, id DESC) AS conversation_rank,
		COUNT(*) OVER (PARTITION BY instance_id, chat_jid) AS message_count,
		SUM(CASE WHEN direction = 'inbound' THEN 1 ELSE 0 END) OVER (PARTITION BY instance_id, chat_jid) AS inbound_count,
		SUM(CASE WHEN direction = 'outbound' THEN 1 ELSE 0 END) OVER (PARTITION BY instance_id, chat_jid) AS outbound_count,
		CASE WHEN direction = 'inbound' THEN 1 ELSE 0 END AS unanswered_inbound
	`)

	query := r.db.WithContext(ctx).
		Table("(?) AS ranked", ranked).
		Select(`
			ranked.instance_id,
			COALESCE(instances.name, '') AS instance_name,
			ranked.chat_jid,
			ranked.source,
			ranked.push_name,
			ranked.is_group,
			ranked.message_id,
			ranked.message_type,
			ranked.text,
			ranked.caption,
			ranked.sent_at,
			ranked.direction,
			ranked.status,
			ranked.message_count,
			ranked.inbound_count,
			ranked.outbound_count,
			ranked.unanswered_inbound
		`).
		Joins("LEFT JOIN instances ON instances.id = ranked.instance_id").
		Where("ranked.conversation_rank = 1")
	if filters.Before != nil {
		query = query.Where("ranked.sent_at < ?", *filters.Before)
	}

	var rows []conversationRow
	if err := query.Order("ranked.sent_at DESC, ranked.chat_jid ASC").Limit(limit + 1).Scan(&rows).Error; err != nil {
		return nil, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]analytics_model.ConversationSummary, 0, len(rows))
	for _, row := range rows {
		preview := strings.TrimSpace(row.Text)
		if preview == "" {
			preview = strings.TrimSpace(row.Caption)
		}
		if preview == "" {
			preview = "[" + row.MessageType + "]"
		}
		items = append(items, analytics_model.ConversationSummary{
			InstanceID:        row.InstanceID,
			InstanceName:      row.InstanceName,
			ChatJID:           row.ChatJID,
			Contact:           conversationContact(row.ChatJID, row.Source),
			PushName:          row.PushName,
			IsGroup:           row.IsGroup,
			LastMessageID:     row.MessageID,
			LastMessageType:   row.MessageType,
			LastMessageText:   preview,
			LastMessageAt:     row.SentAt,
			LastDirection:     row.Direction,
			LastStatus:        row.Status,
			MessageCount:      row.MessageCount,
			InboundCount:      row.InboundCount,
			OutboundCount:     row.OutboundCount,
			UnansweredInbound: row.UnansweredInbound,
		})
	}

	page := &analytics_model.ConversationPage{Items: items, HasMore: hasMore}
	if hasMore && len(items) > 0 {
		page.NextCursor = items[len(items)-1].LastMessageAt.UTC().Format(timeFormat)
	}
	return page, nil
}

func conversationContact(chatJID, source string) string {
	if strings.HasSuffix(chatJID, "@lid") {
		return ""
	}
	return strings.TrimSpace(source)
}

func (r *analyticsRepository) ListMessages(ctx context.Context, filters analytics_model.Filters, chatJID string) (*analytics_model.MessagePage, error) {
	limit := normalizeLimit(filters.Limit)
	query := r.scopedMessages(ctx, filters).Where("chat_jid = ?", chatJID)
	query = applyContentFilters(query, filters)
	if filters.Before != nil {
		query = query.Where("sent_at < ?", *filters.Before)
	}

	var messages []message_model.Message
	if err := query.Order("sent_at DESC, id DESC").Limit(limit + 1).Find(&messages).Error; err != nil {
		return nil, err
	}
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}
	page := &analytics_model.MessagePage{Items: messages, HasMore: hasMore}
	if hasMore && len(messages) > 0 {
		page.NextCursor = messages[len(messages)-1].SentAt.UTC().Format(timeFormat)
	}
	return page, nil
}

func (r *analyticsRepository) scopedMessages(ctx context.Context, filters analytics_model.Filters) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&message_model.Message{}).
		Where("sent_at >= ? AND sent_at <= ?", filters.From, filters.To).
		Where("instance_id IS NOT NULL AND chat_jid <> ''")
	if filters.InstanceID != "" {
		query = query.Where("instance_id = ?", filters.InstanceID)
	}
	return query
}

func applyContentFilters(query *gorm.DB, filters analytics_model.Filters) *gorm.DB {
	if filters.Direction != "" {
		query = query.Where("direction = ?", filters.Direction)
	}
	if filters.MessageType != "" {
		query = query.Where("message_type = ?", filters.MessageType)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.IsGroup != nil {
		query = query.Where("is_group = ?", *filters.IsGroup)
	}
	if filters.Search != "" {
		like := "%" + filters.Search + "%"
		query = query.Where(`
			text ILIKE ? OR caption ILIKE ? OR push_name ILIKE ? OR source ILIKE ? OR chat_jid ILIKE ?
		`, like, like, like, like, like)
	}
	return query
}

func percentage(value, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(value) * 100 / float64(total)
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 100 {
		return 100
	}
	return limit
}

const timeFormat = "2006-01-02T15:04:05.999999999Z07:00"

func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}
