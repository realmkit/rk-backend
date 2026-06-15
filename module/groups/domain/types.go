// Package domain contains group and permission invariants.
package domain

// Key is a stable group key.
type Key string

// Color is a hex UI color.
type Color string

// GroupStatus is the group lifecycle state.
type GroupStatus string

// MembershipStatus is the membership lifecycle state.
type MembershipStatus string

// Action identifies an action that can be granted.
type Action string

// ScopeType identifies the resource type an action applies to.
type ScopeType string

// ObjectType identifies an authorization object type.
type ObjectType = ScopeType

// Relation identifies an authorization relation.
type Relation string

// SubjectType identifies an authorization subject type.
type SubjectType string

// Permission identifies a domain action.
type Permission = Action

// WarningLevel identifies how risky a permission action is.
type WarningLevel string

// ConditionType identifies a supported permission condition.
type ConditionType string

const (
	// WarningLevelNormal marks routine permissions.
	WarningLevelNormal WarningLevel = "normal"

	// WarningLevelSensitive marks permissions that expose private or risky data.
	WarningLevelSensitive WarningLevel = "sensitive"

	// WarningLevelDangerous marks permissions that mutate important state.
	WarningLevelDangerous WarningLevel = "dangerous"
)

const (
	// GroupStatusActive means the group grants permissions.
	GroupStatusActive GroupStatus = "active"

	// GroupStatusDisabled means the group does not grant permissions.
	GroupStatusDisabled GroupStatus = "disabled"

	// GroupStatusSystem means the group is built in and grants permissions.
	GroupStatusSystem GroupStatus = "system"
)

const (
	// MembershipStatusActive means the membership grants permissions.
	MembershipStatusActive MembershipStatus = "active"

	// MembershipStatusDisabled means the membership is disabled.
	MembershipStatusDisabled MembershipStatus = "disabled"

	// MembershipStatusExpired means the membership is expired.
	MembershipStatusExpired MembershipStatus = "expired"

	// MembershipStatusRevoked means the membership is revoked.
	MembershipStatusRevoked MembershipStatus = "revoked"
)

const (
	// ObjectGroup is a group authorization object.
	ObjectGroup ObjectType = "group"

	// ObjectAsset is an asset authorization object.
	ObjectAsset ObjectType = "asset"

	// ObjectUser is a user authorization object.
	ObjectUser ObjectType = "user"

	// ObjectSystem is a system authorization object.
	ObjectSystem ObjectType = "system"

	// ObjectMetadata is a metadata administration authorization object.
	ObjectMetadata ObjectType = "metadata"

	// ObjectForum is a forum authorization object.
	ObjectForum ObjectType = "forum"

	// ObjectForumThread is a forum thread authorization object.
	ObjectForumThread ObjectType = "forum_thread"

	// ObjectForumPost is a forum post authorization object.
	ObjectForumPost ObjectType = "forum_post"

	// ObjectEvent is an event administration authorization object.
	ObjectEvent ObjectType = "event"

	// ObjectCronJob is a cron job authorization object.
	ObjectCronJob ObjectType = "cronjob"

	// ObjectPunishment is a punishment authorization object.
	ObjectPunishment ObjectType = "punishment"

	// ObjectTicket is a ticket authorization object.
	ObjectTicket ObjectType = "ticket"
)

const (
	// SubjectUser is a user subject.
	SubjectUser SubjectType = "user"

	// SubjectGroup is a group subject.
	SubjectGroup SubjectType = "group"

	// SubjectPublic grants access to anonymous and authenticated actors.
	SubjectPublic SubjectType = "public"

	// SubjectAuthenticated grants access to any authenticated local user.
	SubjectAuthenticated SubjectType = "authenticated"
)

const (
	// RelationMember is a member relation.
	RelationMember Relation = "member"

	// RelationViewer is a viewer relation.
	RelationViewer Relation = "viewer"

	// RelationEditor is an editor relation.
	RelationEditor Relation = "editor"

	// RelationManager is a manager relation.
	RelationManager Relation = "manager"

	// RelationOwner is an owner relation.
	RelationOwner Relation = "owner"

	// RelationSelf is a self relation.
	RelationSelf Relation = "self"

	// RelationCreator is a creator relation.
	RelationCreator Relation = "creator"

	// RelationReplyer is a replyer relation.
	RelationReplyer Relation = "replyer"

	// RelationLiker is a liker relation.
	RelationLiker Relation = "liker"

	// RelationModerator is a moderator relation.
	RelationModerator Relation = "moderator"

	// RelationAuthor is an author relation.
	RelationAuthor Relation = "author"

	// RelationIssuer is an issuer relation.
	RelationIssuer Relation = "issuer"

	// RelationTarget is a target relation.
	RelationTarget Relation = "target"

	// RelationSubmitter is the user who opened a ticket.
	RelationSubmitter Relation = "submitter"

	// RelationAssignee is the staff user assigned to a ticket.
	RelationAssignee Relation = "assignee"

	// RelationTeamMember is a group member handling a ticket queue.
	RelationTeamMember Relation = "team_member"
)

const (
	// ConditionEquals requires a context field to equal one value.
	ConditionEquals ConditionType = "equals"

	// ConditionIn requires a context field to match one of many values.
	ConditionIn ConditionType = "in"

	// ConditionFieldEqualsActor requires a context field to equal the actor user ID.
	ConditionFieldEqualsActor ConditionType = "field_equals_actor"

	// ConditionFieldNotEqualsActor requires a context field to differ from the actor user ID.
	ConditionFieldNotEqualsActor ConditionType = "field_not_equals_actor"

	// ConditionIsUnset requires a context field to be missing or empty.
	ConditionIsUnset ConditionType = "is_unset"

	// ConditionAssignedToActor requires an assignment field to equal the actor user ID.
	ConditionAssignedToActor ConditionType = "assigned_to_actor"

	// ConditionWithinDuration requires a timestamp field to be within a duration from now.
	ConditionWithinDuration ConditionType = "within_duration"

	// ConditionOlderThan requires a timestamp field to be older than a duration from now.
	ConditionOlderThan ConditionType = "older_than"
)
