package domain

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

	// PermissionPunishmentsManageIntegrations allows managing integrations.
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

	// PermissionTicketsEscalate allows escalating tickets.
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
