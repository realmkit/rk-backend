package domain

import (
	"regexp"
	"strings"
)

// EventKey identifies one event fact type.
type EventKey string

// Producer identifies the package or module that published an event.
type Producer string

// AggregateType identifies the aggregate affected by an event.
type AggregateType string

// Status is the durable event dispatch status.
type Status string

// ScopeType identifies one event audience scope.
type ScopeType string

const (
	// ProducerUsers identifies user/auth events.
	ProducerUsers Producer = "users"

	// ProducerAssets identifies asset events.
	ProducerAssets Producer = "assets"

	// ProducerMetadata identifies metadata events.
	ProducerMetadata Producer = "metadata"

	// ProducerGroups identifies groups events.
	ProducerGroups Producer = "groups"

	// ProducerForums identifies forums events.
	ProducerForums Producer = "forums"

	// ProducerPunishments identifies punishment events.
	ProducerPunishments Producer = "punishments"

	// ProducerTickets identifies ticket and appeal events.
	ProducerTickets Producer = "tickets"

	// ProducerCronjob identifies cronjob events.
	ProducerCronjob Producer = "cronjob"

	// ProducerNotifications identifies notification events.
	ProducerNotifications Producer = "notifications"

	// ProducerMessages identifies messaging events.
	ProducerMessages Producer = "messages"
)

const (
	// EventUsersUserProvisioned is emitted when a user is provisioned.
	EventUsersUserProvisioned EventKey = "users.user.provisioned"

	// EventAssetsAssetUploadCompleted is emitted when an asset upload finishes.
	EventAssetsAssetUploadCompleted EventKey = "assets.asset.upload_completed"

	// EventMetadataMetafieldSet is emitted when a metafield is set.
	EventMetadataMetafieldSet EventKey = "metadata.metafield.set"

	// EventGroupsMembershipAdded is emitted when group membership is added.
	EventGroupsMembershipAdded EventKey = "groups.membership.added"

	// EventForumsThreadCreated is emitted when a forum thread is created.
	EventForumsThreadCreated EventKey = "forums.thread.created"

	// EventForumsPostCreated is emitted when a forum post is created.
	EventForumsPostCreated EventKey = "forums.post.created"

	// EventPunishmentsPunishmentIssued is emitted when a punishment is issued.
	EventPunishmentsPunishmentIssued EventKey = "punishments.punishment.issued"

	// EventTicketsTicketCreated is emitted when a ticket is opened.
	EventTicketsTicketCreated EventKey = "tickets.ticket.created"

	// EventTicketsMessageCreated is emitted when a ticket message is added.
	EventTicketsMessageCreated EventKey = "tickets.message.created"

	// EventCronjobRunCompleted is emitted when a cron job run succeeds.
	EventCronjobRunCompleted EventKey = "cronjob.run.completed"

	// EventNotificationsNotificationCreated is emitted for a notification.
	EventNotificationsNotificationCreated EventKey = "notifications.notification.created"

	// EventMessagesMessageSent is emitted when a message is sent.
	EventMessagesMessageSent EventKey = "messages.message.sent"
)

const (
	// StatusPending means the event is ready to be claimed.
	StatusPending Status = "pending"

	// StatusProcessing means a dispatcher currently owns the event.
	StatusProcessing Status = "processing"

	// StatusProcessed means dispatch completed.
	StatusProcessed Status = "processed"

	// StatusFailed means dispatch failed and can be retried.
	StatusFailed Status = "failed"

	// StatusDead means retry attempts are exhausted.
	StatusDead Status = "dead"

	// StatusCancelled means an operator cancelled the event.
	StatusCancelled Status = "cancelled"
)

const (
	// ScopeGlobal addresses public global subscribers.
	ScopeGlobal ScopeType = "global"

	// ScopeUser addresses one authenticated user.
	ScopeUser ScopeType = "user"

	// ScopeGroup addresses members of one group.
	ScopeGroup ScopeType = "group"

	// ScopePermission addresses actors with one permission.
	ScopePermission ScopeType = "permission"

	// ScopeForum addresses one forum.
	ScopeForum ScopeType = "forum"

	// ScopeThread addresses one forum thread.
	ScopeThread ScopeType = "thread"

	// ScopePost addresses one forum post.
	ScopePost ScopeType = "post"

	// ScopeAsset addresses one asset.
	ScopeAsset ScopeType = "asset"

	// ScopePunishment addresses one punishment.
	ScopePunishment ScopeType = "punishment"

	// ScopeTicket addresses one ticket.
	ScopeTicket ScopeType = "ticket"

	// ScopeStaff addresses staff-only subscribers.
	ScopeStaff ScopeType = "staff"

	// ScopeSystem addresses backend consumers only.
	ScopeSystem ScopeType = "system"
)

// eventKeyPattern stores package state.
var eventKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)

// ValidateEventKey validates a dotted event key.
func ValidateEventKey(field string, value EventKey) []Violation {
	trimmed := strings.TrimSpace(string(value))
	if trimmed == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	if !eventKeyPattern.MatchString(trimmed) {
		return []Violation{{Field: field, Message: "must be lower dotted words"}}
	}
	return nil
}

// ValidateStatus validates an event status.
func ValidateStatus(field string, value Status) []Violation {
	switch value {
	case StatusPending, StatusProcessing, StatusProcessed,
		StatusFailed, StatusDead, StatusCancelled:
		return nil
	default:
		return []Violation{{Field: field, Message: "is not supported"}}
	}
}

// ValidateScopeType validates a scope type.
func ValidateScopeType(field string, value ScopeType) []Violation {
	switch value {
	case ScopeGlobal, ScopeUser, ScopeGroup, ScopePermission, ScopeForum,
		ScopeThread, ScopePost, ScopeAsset, ScopePunishment, ScopeTicket, ScopeStaff,
		ScopeSystem:
		return nil
	default:
		return []Violation{{Field: field, Message: "is not supported"}}
	}
}
