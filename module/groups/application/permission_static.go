package application

import "github.com/realmkit/rk-backend/module/groups/domain"

// staticPermissionRule defines built-in fallback policy rules.
type staticPermissionRule struct {
	objectType domain.ObjectType
	relations  []domain.Relation
}

// staticPermissionRules maps permissions to built-in fallback requirements.
var staticPermissionRules = map[domain.Permission]staticPermissionRule{
	"groups.read":                                  static(domain.ObjectGroup, groupReaders),
	"groups.update":                                static(domain.ObjectGroup, managers),
	"groups.delete":                                static(domain.ObjectGroup, owners),
	"groups.assign_member":                         static(domain.ObjectGroup, managers),
	"groups.read_members":                          static(domain.ObjectGroup, groupReaders),
	"assets.view":                                  static(domain.ObjectAsset, viewers),
	"assets.update":                                static(domain.ObjectAsset, editors),
	"metadata.write_user":                          static(domain.ObjectUser, selfManagers),
	domain.PermissionForumsView:                    static(domain.ObjectForum, viewers),
	domain.PermissionForumsManageForum:             static(domain.ObjectForum, managers),
	domain.PermissionForumsCreateThread:            static(domain.ObjectForum, creators),
	domain.PermissionForumsReply:                   static(domain.ObjectForum, replyers),
	domain.PermissionForumsLikePosts:               static(domain.ObjectForum, likers),
	domain.PermissionForumsPinThreads:              static(domain.ObjectForum, moderators),
	domain.PermissionForumsManageThreads:           static(domain.ObjectForum, moderators),
	domain.PermissionForumsManagePosts:             static(domain.ObjectForum, moderators),
	domain.PermissionThreadsView:                   static(domain.ObjectForumThread, threadReaders),
	domain.PermissionThreadsUpdate:                 static(domain.ObjectForumThread, threadEditors),
	domain.PermissionThreadsClose:                  static(domain.ObjectForumThread, moderators),
	domain.PermissionThreadsOpen:                   static(domain.ObjectForumThread, moderators),
	domain.PermissionThreadsDelete:                 static(domain.ObjectForumThread, authorModerators),
	domain.PermissionThreadsPin:                    static(domain.ObjectForumThread, moderators),
	domain.PermissionPostsView:                     static(domain.ObjectForumPost, postReaders),
	domain.PermissionPostsUpdate:                   static(domain.ObjectForumPost, postEditors),
	domain.PermissionPostsDelete:                   static(domain.ObjectForumPost, authorModerators),
	domain.PermissionPostsLike:                     static(domain.ObjectForumPost, postLikers),
	domain.PermissionPostsViewHidden:               static(domain.ObjectForumPost, moderators),
	domain.PermissionPostsViewRevisions:            static(domain.ObjectForumPost, moderators),
	domain.PermissionPunishmentsView:               static(domain.ObjectPunishment, punishmentReaders),
	domain.PermissionPunishmentsViewPrivate:        static(domain.ObjectPunishment, moderators),
	domain.PermissionPunishmentsIssue:              static(domain.ObjectPunishment, moderators),
	domain.PermissionPunishmentsRevoke:             static(domain.ObjectPunishment, moderators),
	domain.PermissionPunishmentsUpdate:             static(domain.ObjectPunishment, moderators),
	domain.PermissionPunishmentsManageDefinitions:  static(domain.ObjectPunishment, managers),
	domain.PermissionPunishmentsManageIntegrations: static(domain.ObjectPunishment, managers),
	domain.PermissionPunishmentsViewEvents:         static(domain.ObjectPunishment, moderators),
	domain.PermissionPunishmentsReplayEvents:       static(domain.ObjectPunishment, managers),
	domain.PermissionTicketsView:                   static(domain.ObjectTicket, ticketReaders),
	domain.PermissionTicketsViewPrivate:            static(domain.ObjectTicket, ticketStaff),
	domain.PermissionTicketsCreate:                 static(domain.ObjectTicket, creators),
	domain.PermissionTicketsReply:                  static(domain.ObjectTicket, ticketReplyers),
	domain.PermissionTicketsReplyStaffOnly:         static(domain.ObjectTicket, ticketStaff),
	domain.PermissionTicketsAddEvidence:            static(domain.ObjectTicket, ticketEvidenceEditors),
	domain.PermissionTicketsAssign:                 static(domain.ObjectTicket, ticketManagers),
	domain.PermissionTicketsEscalate:               static(domain.ObjectTicket, ticketStaff),
	domain.PermissionTicketsClose:                  static(domain.ObjectTicket, ticketClosers),
	domain.PermissionTicketsReopen:                 static(domain.ObjectTicket, ticketStaff),
	domain.PermissionTicketsManage:                 static(domain.ObjectTicket, ticketManagers),
	domain.PermissionTicketsManageDefinitions:      static(domain.ObjectTicket, managers),
	domain.PermissionTicketsPerformActions:         static(domain.ObjectTicket, assignedModerators),
	domain.PermissionTicketsAcceptAppeal:           static(domain.ObjectTicket, ticketStaff),
	domain.PermissionTicketsRejectAppeal:           static(domain.ObjectTicket, ticketStaff),
	domain.PermissionTicketsLinkPunishment:         static(domain.ObjectTicket, ticketStaff),
}

var (
	owners           = with(domain.RelationOwner)
	managers         = with(domain.RelationManager, domain.RelationOwner)
	viewers          = with(domain.RelationViewer, domain.RelationManager, domain.RelationOwner)
	editors          = with(domain.RelationEditor, domain.RelationOwner)
	selfManagers     = with(domain.RelationSelf, domain.RelationManager)
	groupReaders     = with(domain.RelationViewer, domain.RelationManager, domain.RelationMember, domain.RelationOwner)
	creators         = with(domain.RelationCreator, domain.RelationManager, domain.RelationOwner)
	replyers         = with(domain.RelationReplyer, domain.RelationManager, domain.RelationOwner)
	likers           = with(domain.RelationLiker, domain.RelationManager, domain.RelationOwner)
	moderators       = with(domain.RelationModerator, domain.RelationManager, domain.RelationOwner)
	authorModerators = with(
		domain.RelationAuthor,
		domain.RelationModerator,
		domain.RelationManager,
		domain.RelationOwner,
	)
	threadReaders = with(
		domain.RelationViewer,
		domain.RelationAuthor,
		domain.RelationModerator,
		domain.RelationManager,
		domain.RelationOwner,
	)
	threadEditors = with(
		domain.RelationAuthor,
		domain.RelationEditor,
		domain.RelationModerator,
		domain.RelationManager,
		domain.RelationOwner,
	)
	postReaders = threadReaders
	postEditors = threadEditors
	postLikers  = with(
		domain.RelationLiker,
		domain.RelationViewer,
		domain.RelationAuthor,
		domain.RelationModerator,
		domain.RelationManager,
		domain.RelationOwner,
	)
	punishmentReaders = with(
		domain.RelationViewer,
		domain.RelationTarget,
		domain.RelationIssuer,
		domain.RelationModerator,
		domain.RelationManager,
		domain.RelationOwner,
	)
	ticketStaff = with(
		domain.RelationAssignee,
		domain.RelationTeamMember,
		domain.RelationModerator,
		domain.RelationManager,
		domain.RelationOwner,
	)
	ticketManagers = with(
		domain.RelationTeamMember,
		domain.RelationModerator,
		domain.RelationManager,
		domain.RelationOwner,
	)
	assignedModerators = with(
		domain.RelationAssignee,
		domain.RelationModerator,
		domain.RelationManager,
		domain.RelationOwner,
	)
	ticketReaders = with(
		domain.RelationSubmitter,
		domain.RelationAssignee,
		domain.RelationTeamMember,
		domain.RelationViewer,
		domain.RelationModerator,
		domain.RelationManager,
		domain.RelationOwner,
	)
	ticketReplyers = with(
		domain.RelationSubmitter,
		domain.RelationAssignee,
		domain.RelationTeamMember,
		domain.RelationReplyer,
		domain.RelationModerator,
		domain.RelationManager,
		domain.RelationOwner,
	)
	ticketEvidenceEditors = with(
		domain.RelationSubmitter,
		domain.RelationAssignee,
		domain.RelationTeamMember,
		domain.RelationEditor,
		domain.RelationModerator,
		domain.RelationManager,
		domain.RelationOwner,
	)
	ticketClosers = with(
		domain.RelationSubmitter,
		domain.RelationAssignee,
		domain.RelationTeamMember,
		domain.RelationModerator,
		domain.RelationManager,
		domain.RelationOwner,
	)
)

// static creates one static permission rule.
func static(objectType domain.ObjectType, relations []domain.Relation) staticPermissionRule {
	return staticPermissionRule{objectType: objectType, relations: relations}
}

// with creates a relation set.
func with(relations ...domain.Relation) []domain.Relation {
	return relations
}
