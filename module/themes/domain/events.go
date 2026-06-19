package domain

// EventKey identifies one theme event fact.
type EventKey string

const (
	// EventThemeCreated is emitted when a theme family is created.
	EventThemeCreated EventKey = "themes.theme.created"
	// EventThemeUpdated is emitted when theme metadata changes.
	EventThemeUpdated EventKey = "themes.theme.updated"
	// EventVersionImported is emitted when a package import is persisted.
	EventVersionImported EventKey = "themes.version.imported"
	// EventVersionValidated is emitted when validation completes.
	EventVersionValidated EventKey = "themes.version.validated"
	// EventVersionFileSaved is emitted when a draft file changes.
	EventVersionFileSaved EventKey = "themes.version.file_saved"
	// EventVersionArchived is emitted when a version is archived.
	EventVersionArchived EventKey = "themes.version.archived"
	// EventActivationChanged is emitted when public or preview activation changes.
	EventActivationChanged EventKey = "themes.activation.changed"
	// EventActivationRolledBack is emitted when rollback creates a new activation.
	EventActivationRolledBack EventKey = "themes.activation.rolled_back"
	// EventSigningKeyCreated is emitted when an operator adds a signing key.
	EventSigningKeyCreated EventKey = "themes.signing_key.created"
	// EventSigningKeyRetired is emitted when a key is retired.
	EventSigningKeyRetired EventKey = "themes.signing_key.retired"
	// EventSigningKeyRevoked is emitted when a key is revoked.
	EventSigningKeyRevoked EventKey = "themes.signing_key.revoked"
	// EventCacheInvalidated is emitted when theme delivery caches must refresh.
	EventCacheInvalidated EventKey = "themes.cache.invalidated"
)

// ThemeEventKeys returns all built-in theme event keys.
func ThemeEventKeys() []EventKey {
	return []EventKey{
		EventThemeCreated,
		EventThemeUpdated,
		EventVersionImported,
		EventVersionValidated,
		EventVersionFileSaved,
		EventVersionArchived,
		EventActivationChanged,
		EventActivationRolledBack,
		EventSigningKeyCreated,
		EventSigningKeyRetired,
		EventSigningKeyRevoked,
		EventCacheInvalidated,
	}
}
