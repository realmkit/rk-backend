package domain

// Descriptor describes one supported event type.
type Descriptor struct {
	// Key is the stable event key.
	Key EventKey `json:"key"`

	// SchemaVersion is the current schema version.
	SchemaVersion int `json:"schema_version"`

	// Producer is the publishing module or package.
	Producer Producer `json:"producer"`

	// AggregateType is the affected aggregate type.
	AggregateType AggregateType `json:"aggregate_type"`

	// Private reports whether payloads are staff/system only by default.
	Private bool `json:"private"`
}

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

	// EventCronjobRunCompleted is emitted when a cron job run succeeds.
	EventCronjobRunCompleted EventKey = "cronjob.run.completed"

	// EventNotificationsNotificationCreated is emitted for a notification.
	EventNotificationsNotificationCreated EventKey = "notifications.notification.created"

	// EventMessagesMessageSent is emitted when a message is sent.
	EventMessagesMessageSent EventKey = "messages.message.sent"
)

// Catalog returns descriptors for known GameHub events.
func Catalog() []Descriptor {
	return []Descriptor{
		{Key: "users.user.provisioned", SchemaVersion: 1, Producer: ProducerUsers, AggregateType: "user"},
		{Key: "users.user.updated", SchemaVersion: 1, Producer: ProducerUsers, AggregateType: "user"},
		{Key: "users.identity.linked", SchemaVersion: 1, Producer: ProducerUsers, AggregateType: "identity"},
		{Key: "users.identity.unlinked", SchemaVersion: 1, Producer: ProducerUsers, AggregateType: "identity"},
		{Key: "users.identity.claim_refreshed", SchemaVersion: 1, Producer: ProducerUsers, AggregateType: "identity"},
		{Key: "assets.asset.created", SchemaVersion: 1, Producer: ProducerAssets, AggregateType: "asset"},
		{Key: "assets.asset.upload_completed", SchemaVersion: 1, Producer: ProducerAssets, AggregateType: "asset"},
		{Key: "assets.asset.updated", SchemaVersion: 1, Producer: ProducerAssets, AggregateType: "asset"},
		{Key: "assets.asset.deleted", SchemaVersion: 1, Producer: ProducerAssets, AggregateType: "asset"},
		{Key: "assets.upload_intent.expired", SchemaVersion: 1, Producer: ProducerAssets, AggregateType: "asset"},
		{Key: "metadata.definition.created", SchemaVersion: 1, Producer: ProducerMetadata, AggregateType: "metadata_definition"},
		{Key: "metadata.definition.updated", SchemaVersion: 1, Producer: ProducerMetadata, AggregateType: "metadata_definition"},
		{Key: "metadata.definition.deleted", SchemaVersion: 1, Producer: ProducerMetadata, AggregateType: "metadata_definition"},
		{Key: "metadata.entry.created", SchemaVersion: 1, Producer: ProducerMetadata, AggregateType: "metadata_entry"},
		{Key: "metadata.entry.updated", SchemaVersion: 1, Producer: ProducerMetadata, AggregateType: "metadata_entry"},
		{Key: "metadata.entry.deleted", SchemaVersion: 1, Producer: ProducerMetadata, AggregateType: "metadata_entry"},
		{Key: "metadata.metafield.set", SchemaVersion: 1, Producer: ProducerMetadata, AggregateType: "metafield"},
		{Key: "metadata.metafield.deleted", SchemaVersion: 1, Producer: ProducerMetadata, AggregateType: "metafield"},
		{Key: "groups.group.created", SchemaVersion: 1, Producer: ProducerGroups, AggregateType: "group"},
		{Key: "groups.group.updated", SchemaVersion: 1, Producer: ProducerGroups, AggregateType: "group"},
		{Key: "groups.group.deleted", SchemaVersion: 1, Producer: ProducerGroups, AggregateType: "group"},
		{Key: "groups.membership.added", SchemaVersion: 1, Producer: ProducerGroups, AggregateType: "group_membership"},
		{Key: "groups.membership.removed", SchemaVersion: 1, Producer: ProducerGroups, AggregateType: "group_membership"},
		{Key: "groups.relation_tuple.created", SchemaVersion: 1, Producer: ProducerGroups, AggregateType: "relation_tuple"},
		{Key: "groups.relation_tuple.deleted", SchemaVersion: 1, Producer: ProducerGroups, AggregateType: "relation_tuple"},
		{Key: "groups.permission.policy_changed", SchemaVersion: 1, Producer: ProducerGroups, AggregateType: "permission_policy", Private: true},
		{Key: "forums.category.created", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_category"},
		{Key: "forums.category.updated", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_category"},
		{Key: "forums.category.deleted", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_category"},
		{Key: "forums.forum.created", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum"},
		{Key: "forums.forum.updated", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum"},
		{Key: "forums.forum.moved", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum"},
		{Key: "forums.forum.deleted", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum"},
		{Key: "forums.forum.settings_updated", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum"},
		{Key: "forums.forum.permissions_updated", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum", Private: true},
		{Key: "forums.thread.created", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_thread"},
		{Key: "forums.thread.updated", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_thread"},
		{Key: "forums.thread.deleted", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_thread"},
		{Key: "forums.thread.closed", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_thread"},
		{Key: "forums.thread.opened", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_thread"},
		{Key: "forums.thread.locked", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_thread"},
		{Key: "forums.thread.pinned", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_thread"},
		{Key: "forums.thread.unpinned", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_thread"},
		{Key: "forums.thread.moved", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_thread"},
		{Key: "forums.post.created", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_post"},
		{Key: "forums.post.updated", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_post"},
		{Key: "forums.post.deleted", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_post"},
		{Key: "forums.post.hidden", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_post"},
		{Key: "forums.post.restored", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_post"},
		{Key: "forums.post.liked", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_post"},
		{Key: "forums.post.unliked", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_post"},
		{Key: "forums.thread.read", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_thread"},
		{Key: "forums.forum.read", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum"},
		{Key: "forums.stats.rebuilt", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_stats", Private: true},
		{Key: "forums.likes.rebuilt", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_stats", Private: true},
		{Key: "forums.views.flushed", SchemaVersion: 1, Producer: ProducerForums, AggregateType: "forum_thread", Private: true},
		{Key: "punishments.definition.created", SchemaVersion: 1, Producer: ProducerPunishments, AggregateType: "punishment_definition", Private: true},
		{Key: "punishments.definition.updated", SchemaVersion: 1, Producer: ProducerPunishments, AggregateType: "punishment_definition", Private: true},
		{Key: "punishments.definition.deleted", SchemaVersion: 1, Producer: ProducerPunishments, AggregateType: "punishment_definition", Private: true},
		{Key: "punishments.punishment.issued", SchemaVersion: 1, Producer: ProducerPunishments, AggregateType: "punishment", Private: true},
		{Key: "punishments.punishment.updated", SchemaVersion: 1, Producer: ProducerPunishments, AggregateType: "punishment", Private: true},
		{Key: "punishments.punishment.revoked", SchemaVersion: 1, Producer: ProducerPunishments, AggregateType: "punishment", Private: true},
		{Key: "punishments.punishment.expired", SchemaVersion: 1, Producer: ProducerPunishments, AggregateType: "punishment", Private: true},
		{Key: "punishments.restrictions.rebuilt", SchemaVersion: 1, Producer: ProducerPunishments, AggregateType: "punishment_restriction", Private: true},
		{Key: "punishments.integration.requested", SchemaVersion: 1, Producer: ProducerPunishments, AggregateType: "punishment", Private: true},
		{Key: "punishments.integration.completed", SchemaVersion: 1, Producer: ProducerPunishments, AggregateType: "punishment", Private: true},
		{Key: "punishments.integration.failed", SchemaVersion: 1, Producer: ProducerPunishments, AggregateType: "punishment", Private: true},
		{Key: "cronjob.run.started", SchemaVersion: 1, Producer: ProducerCronjob, AggregateType: "cronjob_run", Private: true},
		{Key: "cronjob.run.completed", SchemaVersion: 1, Producer: ProducerCronjob, AggregateType: "cronjob_run", Private: true},
		{Key: "cronjob.run.failed", SchemaVersion: 1, Producer: ProducerCronjob, AggregateType: "cronjob_run", Private: true},
		{Key: "cronjob.run.skipped", SchemaVersion: 1, Producer: ProducerCronjob, AggregateType: "cronjob_run", Private: true},
		{Key: "cronjob.definition.updated", SchemaVersion: 1, Producer: ProducerCronjob, AggregateType: "cronjob_definition", Private: true},
		{Key: "notifications.notification.created", SchemaVersion: 1, Producer: ProducerNotifications, AggregateType: "notification"},
		{Key: "notifications.notification.read", SchemaVersion: 1, Producer: ProducerNotifications, AggregateType: "notification"},
		{Key: "notifications.notification.deleted", SchemaVersion: 1, Producer: ProducerNotifications, AggregateType: "notification"},
		{Key: "notifications.digest.sent", SchemaVersion: 1, Producer: ProducerNotifications, AggregateType: "notification_digest"},
		{Key: "messages.conversation.created", SchemaVersion: 1, Producer: ProducerMessages, AggregateType: "conversation"},
		{Key: "messages.message.sent", SchemaVersion: 1, Producer: ProducerMessages, AggregateType: "message"},
		{Key: "messages.message.edited", SchemaVersion: 1, Producer: ProducerMessages, AggregateType: "message"},
		{Key: "messages.message.deleted", SchemaVersion: 1, Producer: ProducerMessages, AggregateType: "message"},
		{Key: "messages.message.read", SchemaVersion: 1, Producer: ProducerMessages, AggregateType: "message"},
		{Key: "messages.typing.started", SchemaVersion: 1, Producer: ProducerMessages, AggregateType: "conversation"},
		{Key: "messages.typing.stopped", SchemaVersion: 1, Producer: ProducerMessages, AggregateType: "conversation"},
	}
}
