package domain

import (
	"regexp"
	"slices"
	"strings"
)

// Key is a stable group key.
type Key string

// Color is a hex UI color.
type Color string

// GroupStatus is the group lifecycle state.
type GroupStatus string

// MembershipStatus is the membership lifecycle state.
type MembershipStatus string

// ObjectType identifies an authorization object type.
type ObjectType string

// Relation identifies an authorization relation.
type Relation string

// SubjectType identifies an authorization subject type.
type SubjectType string

// Permission identifies a domain action.
type Permission string

// ConditionType identifies a supported permission condition.
type ConditionType string

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
	// PermissionForumsView allows reading a forum.
	PermissionForumsView Permission = "forums.view"

	// PermissionForumsManageForum allows managing forum structure and settings.
	PermissionForumsManageForum Permission = "forums.manage_forum"

	// PermissionForumsCreateThread allows creating threads inside a forum.
	PermissionForumsCreateThread Permission = "forums.create_thread"

	// PermissionForumsReply allows replying to threads inside a forum.
	PermissionForumsReply Permission = "forums.reply"

	// PermissionForumsLikePosts allows liking posts inside a forum.
	PermissionForumsLikePosts Permission = "forums.like_posts"

	// PermissionForumsPinThreads allows pinning threads inside a forum.
	PermissionForumsPinThreads Permission = "forums.pin_threads"

	// PermissionForumsManageThreads allows moderating threads inside a forum.
	PermissionForumsManageThreads Permission = "forums.manage_threads"

	// PermissionForumsManagePosts allows moderating posts inside a forum.
	PermissionForumsManagePosts Permission = "forums.manage_posts"

	// PermissionThreadsView allows reading a thread.
	PermissionThreadsView Permission = "threads.view"

	// PermissionThreadsUpdate allows updating a thread.
	PermissionThreadsUpdate Permission = "threads.update"

	// PermissionThreadsClose allows closing a thread.
	PermissionThreadsClose Permission = "threads.close"

	// PermissionThreadsOpen allows opening a thread.
	PermissionThreadsOpen Permission = "threads.open"

	// PermissionThreadsDelete allows deleting a thread.
	PermissionThreadsDelete Permission = "threads.delete"

	// PermissionThreadsPin allows pinning a thread.
	PermissionThreadsPin Permission = "threads.pin"

	// PermissionPostsView allows reading a post.
	PermissionPostsView Permission = "posts.view"

	// PermissionPostsUpdate allows updating a post.
	PermissionPostsUpdate Permission = "posts.update"

	// PermissionPostsDelete allows deleting a post.
	PermissionPostsDelete Permission = "posts.delete"

	// PermissionPostsLike allows liking a post.
	PermissionPostsLike Permission = "posts.like"

	// PermissionPostsViewHidden allows reading hidden posts.
	PermissionPostsViewHidden Permission = "posts.view_hidden"

	// PermissionPostsViewRevisions allows reading post revisions.
	PermissionPostsViewRevisions Permission = "posts.view_revisions"

	// PermissionEventsView allows reading event outbox diagnostics.
	PermissionEventsView Permission = "events.view"

	// PermissionEventsReplay allows replaying or cancelling events.
	PermissionEventsReplay Permission = "events.replay"

	// PermissionCronJobsView allows reading cron job status and history.
	PermissionCronJobsView Permission = "cronjobs.view"

	// PermissionCronJobsManage allows changing cron job schedules.
	PermissionCronJobsManage Permission = "cronjobs.manage"

	// PermissionCronJobsRun allows manually running cron jobs.
	PermissionCronJobsRun Permission = "cronjobs.run"

	// PermissionCronJobsRepair allows repairing cron job locks.
	PermissionCronJobsRepair Permission = "cronjobs.repair"

	// PermissionPunishmentsView allows reading punishments.
	PermissionPunishmentsView Permission = "punishments.view"

	// PermissionPunishmentsViewPrivate allows reading private punishment fields.
	PermissionPunishmentsViewPrivate Permission = "punishments.view_private"

	// PermissionPunishmentsIssue allows issuing punishments.
	PermissionPunishmentsIssue Permission = "punishments.issue"

	// PermissionPunishmentsRevoke allows revoking punishments.
	PermissionPunishmentsRevoke Permission = "punishments.revoke"

	// PermissionPunishmentsUpdate allows updating punishments.
	PermissionPunishmentsUpdate Permission = "punishments.update"

	// PermissionPunishmentsManageDefinitions allows managing punishment definitions.
	PermissionPunishmentsManageDefinitions Permission = "punishments.manage_definitions"

	// PermissionPunishmentsManageIntegrations allows managing punishment integrations.
	PermissionPunishmentsManageIntegrations Permission = "punishments.manage_integrations"

	// PermissionPunishmentsViewEvents allows reading punishment events.
	PermissionPunishmentsViewEvents Permission = "punishments.view_events"

	// PermissionPunishmentsReplayEvents allows replaying punishment events.
	PermissionPunishmentsReplayEvents Permission = "punishments.replay_events"

	// PermissionTicketsView allows reading tickets.
	PermissionTicketsView Permission = "tickets.view"

	// PermissionTicketsViewPrivate allows reading staff-only ticket content.
	PermissionTicketsViewPrivate Permission = "tickets.view_private"

	// PermissionTicketsCreate allows opening tickets.
	PermissionTicketsCreate Permission = "tickets.create"

	// PermissionTicketsReply allows replying to tickets.
	PermissionTicketsReply Permission = "tickets.reply"

	// PermissionTicketsReplyStaffOnly allows adding staff-only ticket messages.
	PermissionTicketsReplyStaffOnly Permission = "tickets.reply_staff_only"

	// PermissionTicketsAddEvidence allows adding ticket evidence.
	PermissionTicketsAddEvidence Permission = "tickets.add_evidence"

	// PermissionTicketsAssign allows assigning tickets.
	PermissionTicketsAssign Permission = "tickets.assign"

	// PermissionTicketsEscalate allows escalating tickets to another team.
	PermissionTicketsEscalate Permission = "tickets.escalate"

	// PermissionTicketsClose allows closing tickets.
	PermissionTicketsClose Permission = "tickets.close"

	// PermissionTicketsReopen allows reopening tickets.
	PermissionTicketsReopen Permission = "tickets.reopen"

	// PermissionTicketsManage allows managing ticket queues.
	PermissionTicketsManage Permission = "tickets.manage"

	// PermissionTicketsManageDefinitions allows managing ticket definitions.
	PermissionTicketsManageDefinitions Permission = "tickets.manage_definitions"

	// PermissionTicketsPerformActions allows executing ticket side effects.
	PermissionTicketsPerformActions Permission = "tickets.perform_actions"

	// PermissionTicketsAcceptAppeal allows accepting punishment appeals.
	PermissionTicketsAcceptAppeal Permission = "tickets.accept_appeal"

	// PermissionTicketsRejectAppeal allows rejecting punishment appeals.
	PermissionTicketsRejectAppeal Permission = "tickets.reject_appeal"

	// PermissionTicketsLinkPunishment allows linking tickets to punishments.
	PermissionTicketsLinkPunishment Permission = "tickets.link_punishment"
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

// keyPattern matches stable lower snake identifiers.
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,62}[a-z0-9]$`)

// colorPattern matches six-digit hex colors.
var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// ValidateKey validates key.
func ValidateKey(field string, key Key) []Violation {
	if !keyPattern.MatchString(strings.TrimSpace(string(key))) {
		return []Violation{{Field: field, Message: "must be lower snake case and between 3 and 64 characters"}}
	}
	return nil
}

// ValidateColor validates color.
func ValidateColor(field string, color Color) []Violation {
	if !colorPattern.MatchString(strings.TrimSpace(string(color))) {
		return []Violation{{Field: field, Message: "must be a hex color"}}
	}
	return nil
}

// ValidateGroupStatus validates group status.
func ValidateGroupStatus(field string, status GroupStatus) []Violation {
	if slices.Contains([]GroupStatus{GroupStatusActive, GroupStatusDisabled, GroupStatusSystem}, status) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateMembershipStatus validates membership status.
func ValidateMembershipStatus(field string, status MembershipStatus) []Violation {
	if slices.Contains([]MembershipStatus{MembershipStatusActive, MembershipStatusDisabled, MembershipStatusExpired, MembershipStatusRevoked}, status) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateRelationTerm validates relation-like lower snake text.
func ValidateRelationTerm(field string, value string) []Violation {
	if !keyPattern.MatchString(strings.TrimSpace(value)) {
		return []Violation{{Field: field, Message: "must be lower snake case and between 3 and 64 characters"}}
	}
	return nil
}

// ValidatePermission validates a dotted permission name.
func ValidatePermission(field string, value Permission) []Violation {
	permission := strings.TrimSpace(string(value))
	if permission == "" || len(permission) > 120 {
		return []Violation{{Field: field, Message: "must be a dotted permission between 1 and 120 characters"}}
	}
	parts := strings.Split(permission, ".")
	for _, part := range parts {
		if !keyPattern.MatchString(part) {
			return []Violation{{Field: field, Message: "must use lower snake case segments separated by dots"}}
		}
	}
	return nil
}

// ValidateConditionType validates condition type.
func ValidateConditionType(field string, conditionType ConditionType) []Violation {
	if slices.Contains([]ConditionType{ConditionEquals, ConditionIn, ConditionFieldEqualsActor, ConditionFieldNotEqualsActor, ConditionIsUnset, ConditionAssignedToActor, ConditionWithinDuration, ConditionOlderThan}, conditionType) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}
