package domain

import (
	"regexp"
	"slices"
	"strings"
)

// OwnerType identifies a RealmKit model that can receive metadata.
type OwnerType string

// Key identifies a metadata field within an owner type.
type Key string

// Handle identifies a metaobject entry.
type Handle string

// MetaobjectType identifies a metaobject definition.
type MetaobjectType string

// ValueType identifies the kind of value accepted by a definition.
type ValueType string

// Supported owner types.
const (
	// OwnerAsset attaches metadata to assets.
	OwnerAsset OwnerType = "asset"

	// OwnerGroup attaches metadata to groups.
	OwnerGroup OwnerType = "group"

	// OwnerUser attaches metadata to users.
	OwnerUser OwnerType = "user"

	// OwnerForumCategory attaches metadata to forum categories.
	OwnerForumCategory OwnerType = "forum_category"

	// OwnerForum attaches metadata to forums.
	OwnerForum OwnerType = "forum"

	// OwnerForumThread attaches metadata to forum threads.
	OwnerForumThread OwnerType = "forum_thread"

	// OwnerPunishmentDefinition attaches metadata to punishment definitions.
	OwnerPunishmentDefinition OwnerType = "punishment_definition"

	// OwnerPunishment attaches metadata to punishments.
	OwnerPunishment OwnerType = "punishment"

	// OwnerTicketDefinition attaches metadata to ticket definitions.
	OwnerTicketDefinition OwnerType = "ticket_definition"

	// OwnerTicket attaches metadata to tickets.
	OwnerTicket OwnerType = "ticket"
)

// Supported value types.
const (
	ValueSingleLineText      ValueType = "single_line_text"
	ValueMultiLineText       ValueType = "multi_line_text"
	ValueInteger             ValueType = "integer"
	ValueDecimal             ValueType = "decimal"
	ValueBoolean             ValueType = "boolean"
	ValueDate                ValueType = "date"
	ValueDatetime            ValueType = "datetime"
	ValueJSON                ValueType = "json"
	ValueURL                 ValueType = "url"
	ValueColor               ValueType = "color"
	ValueEnum                ValueType = "enum"
	ValueOwnerReference      ValueType = "owner_reference"
	ValueMetaobjectReference ValueType = "metaobject_reference"
)

// AllowedOwnerTypes returns the owner types supported by the metadata module.
func AllowedOwnerTypes() []OwnerType {
	return []OwnerType{
		OwnerAsset,
		OwnerGroup,
		OwnerUser,
		OwnerForumCategory,
		OwnerForum,
		OwnerForumThread,
		OwnerPunishmentDefinition,
		OwnerPunishment,
		OwnerTicketDefinition,
		OwnerTicket,
	}
}

// SupportedValueTypes returns the supported metadata value types.
func SupportedValueTypes() []ValueType {
	return []ValueType{
		ValueSingleLineText,
		ValueMultiLineText,
		ValueInteger,
		ValueDecimal,
		ValueBoolean,
		ValueDate,
		ValueDatetime,
		ValueJSON,
		ValueURL,
		ValueColor,
		ValueEnum,
		ValueOwnerReference,
		ValueMetaobjectReference,
	}
}

// ValidOwnerType reports whether ownerType is allowlisted.
func ValidOwnerType(ownerType OwnerType) bool {
	return slices.Contains(AllowedOwnerTypes(), ownerType)
}

// ValidValueType reports whether valueType is supported.
func ValidValueType(valueType ValueType) bool {
	return slices.Contains(SupportedValueTypes(), valueType)
}

// ValidateOwnerType validates ownerType.
func ValidateOwnerType(field string, ownerType OwnerType) []Violation {
	if strings.TrimSpace(string(ownerType)) == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	if !ValidOwnerType(ownerType) {
		return []Violation{{Field: field, Message: "is not supported"}}
	}
	return nil
}

// ValidateKey validates key.
func ValidateKey(field string, key Key) []Violation {
	return validateMachineKey(field, string(key))
}

// ValidateHandle validates handle.
func ValidateHandle(field string, handle Handle) []Violation {
	return validateHandle(field, string(handle))
}

// ValidateMetaobjectType validates objectType.
func ValidateMetaobjectType(field string, objectType MetaobjectType) []Violation {
	return validateMachineKey(field, string(objectType))
}

// machineKeyPattern matches lower snake case identifiers between 3 and 64 characters.
var machineKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,62}[a-z0-9]$`)

// handlePattern matches stable handles between 3 and 120 characters.
var handlePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,118}[a-z0-9]$`)

// validateMachineKey validates lower snake case identifiers.
func validateMachineKey(field string, value string) []Violation {
	value = strings.TrimSpace(value)
	if value == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	if !machineKeyPattern.MatchString(value) {
		return []Violation{{Field: field, Message: "must be lower snake case and between 3 and 64 characters"}}
	}
	return nil
}

// validateHandle validates stable public handles.
func validateHandle(field string, value string) []Violation {
	value = strings.TrimSpace(value)
	if value == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	if !handlePattern.MatchString(value) {
		return []Violation{{Field: field, Message: "must be lower case, stable, and between 3 and 120 characters"}}
	}
	return nil
}
