package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	groupsdomain "github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"github.com/niflaot/gamehub-go/pkg/postgres/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestForumRepositoriesLifecycle verifies category and forum persistence.
func TestForumRepositoriesLifecycle(t *testing.T) {
	categories, forums, _ := newRepositories(t)
	category, err := categories.Create(context.Background(), testCategory())
	if err != nil {
		t.Fatalf("Create category error = %v", err)
	}
	forum := testForum(category.ID, nil, "news")
	created, err := forums.Create(context.Background(), forum)
	if err != nil {
		t.Fatalf("Create forum error = %v", err)
	}
	if created.Path != forum.Path {
		t.Fatalf("created path = %q, want %q", created.Path, forum.Path)
	}
	stats, err := forums.ListStats(context.Background(), []uuid.UUID{created.ID})
	if err != nil {
		t.Fatalf("ListStats() error = %v", err)
	}
	if stats[created.ID].ForumID != created.ID {
		t.Fatalf("stats = %+v, want created stats", stats)
	}
	created.Name = "News Updated"
	updated, err := forums.Update(context.Background(), created, created.Version)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Version != 2 || updated.Name != "News Updated" {
		t.Fatalf("updated = %+v, want version 2", updated)
	}
	list, err := forums.List(context.Background(), port.ForumFilter{CategoryID: category.ID}, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("List() items = %d, want 1", len(list.Items))
	}
}

// TestCategoryRepositoryPreconditionAndSoftDelete verifies versioned deletes.
func TestCategoryRepositoryPreconditionAndSoftDelete(t *testing.T) {
	categories, _, _ := newRepositories(t)
	category, err := categories.Create(context.Background(), testCategory())
	if err != nil {
		t.Fatalf("Create category error = %v", err)
	}

	category.Name = "Changed"
	if _, err := categories.Update(context.Background(), category, category.Version+1); !errors.Is(err, port.ErrPreconditionFailed) {
		t.Fatalf("Update() error = %v, want %v", err, port.ErrPreconditionFailed)
	}
	if err := categories.Delete(context.Background(), category.ID, category.Version); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := categories.FindByID(context.Background(), category.ID); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("FindByID() error = %v, want %v", err, port.ErrNotFound)
	}
}

// TestForumRepositoryMoveUpdatesDescendantPaths verifies materialized path moves.
func TestForumRepositoryMoveUpdatesDescendantPaths(t *testing.T) {
	categories, forums, _ := newRepositories(t)
	category, err := categories.Create(context.Background(), testCategory())
	if err != nil {
		t.Fatalf("Create category error = %v", err)
	}
	parent := testForum(category.ID, nil, "parent")
	parent, err = forums.Create(context.Background(), parent)
	if err != nil {
		t.Fatalf("Create parent error = %v", err)
	}
	child := testForum(category.ID, &parent.ID, "child")
	child.Path = parent.Path + child.ID.String() + "/"
	child.Depth = 1
	child, err = forums.Create(context.Background(), child)
	if err != nil {
		t.Fatalf("Create child error = %v", err)
	}
	oldPath := parent.Path
	parent.Path = "/" + uuid.NewString() + "/" + parent.ID.String() + "/"
	parent.Depth = 1
	moved, err := forums.Move(context.Background(), parent, oldPath, parent.Version)
	if err != nil {
		t.Fatalf("Move() error = %v", err)
	}
	foundChild, err := forums.FindByID(context.Background(), child.ID)
	if err != nil {
		t.Fatalf("Find child error = %v", err)
	}
	if foundChild.Path[:len(moved.Path)] != moved.Path || foundChild.Depth != 2 {
		t.Fatalf("child path=%q depth=%d, want under moved path %q depth 2", foundChild.Path, foundChild.Depth, moved.Path)
	}
}

// TestForumRepositoryReorderPersistsDisplayOrder verifies display order updates persist.
func TestForumRepositoryReorderPersistsDisplayOrder(t *testing.T) {
	categories, forums, _ := newRepositories(t)
	category, err := categories.Create(context.Background(), testCategory())
	if err != nil {
		t.Fatalf("Create category error = %v", err)
	}
	first, err := forums.Create(context.Background(), testForum(category.ID, nil, "first"))
	if err != nil {
		t.Fatalf("Create first forum error = %v", err)
	}
	second, err := forums.Create(context.Background(), testForum(category.ID, nil, "second"))
	if err != nil {
		t.Fatalf("Create second forum error = %v", err)
	}

	if err := forums.Reorder(context.Background(), []port.ReorderItem{{ID: first.ID, DisplayOrder: 20}, {ID: second.ID, DisplayOrder: 10}}); err != nil {
		t.Fatalf("Reorder() error = %v", err)
	}
	foundFirst, err := forums.FindByID(context.Background(), first.ID)
	if err != nil {
		t.Fatalf("Find first error = %v", err)
	}
	foundSecond, err := forums.FindByID(context.Background(), second.ID)
	if err != nil {
		t.Fatalf("Find second error = %v", err)
	}
	if foundFirst.DisplayOrder != 20 || foundSecond.DisplayOrder != 10 {
		t.Fatalf("orders = %d/%d, want 20/10", foundFirst.DisplayOrder, foundSecond.DisplayOrder)
	}
}

// TestVisibilityAuthorizerSupportsPublicAndAuthenticated verifies bulk visibility subjects.
func TestVisibilityAuthorizerSupportsPublicAndAuthenticated(t *testing.T) {
	_, _, db := newRepositories(t)
	store := orm.NewStore(db)
	authorizer := NewVisibilityAuthorizer(store)
	publicForumID := uuid.New()
	authForumID := uuid.New()
	createTuple(t, db, publicForumID, groupsdomain.SubjectPublic, groupsdomain.PublicSubjectID())
	createTuple(t, db, authForumID, groupsdomain.SubjectAuthenticated, groupsdomain.AuthenticatedSubjectID())

	anonymous, err := authorizer.VisibleForums(context.Background(), uuid.Nil, []uuid.UUID{publicForumID, authForumID})
	if err != nil {
		t.Fatalf("VisibleForums anonymous error = %v", err)
	}
	if !anonymous[publicForumID] || anonymous[authForumID] {
		t.Fatalf("anonymous = %+v, want only public", anonymous)
	}
	authenticated, err := authorizer.VisibleForums(context.Background(), uuid.New(), []uuid.UUID{publicForumID, authForumID})
	if err != nil {
		t.Fatalf("VisibleForums authenticated error = %v", err)
	}
	if !authenticated[publicForumID] || !authenticated[authForumID] {
		t.Fatalf("authenticated = %+v, want public and authenticated", authenticated)
	}
}

// TestVisibilityAuthorizerSupportsGroupMembership verifies group member grants.
func TestVisibilityAuthorizerSupportsGroupMembership(t *testing.T) {
	_, _, db := newRepositories(t)
	store := orm.NewStore(db)
	authorizer := NewVisibilityAuthorizer(store)
	forumID := uuid.New()
	groupID := uuid.New()
	memberID := uuid.New()
	otherID := uuid.New()
	createGroup(t, db, groupID)
	createMembership(t, db, groupID, memberID)
	createTuple(t, db, forumID, groupsdomain.SubjectGroup, groupID)

	memberVisible, err := authorizer.VisibleForums(context.Background(), memberID, []uuid.UUID{forumID})
	if err != nil {
		t.Fatalf("VisibleForums member error = %v", err)
	}
	if !memberVisible[forumID] {
		t.Fatalf("memberVisible = %+v, want forum visible", memberVisible)
	}
	otherVisible, err := authorizer.VisibleForums(context.Background(), otherID, []uuid.UUID{forumID})
	if err != nil {
		t.Fatalf("VisibleForums other error = %v", err)
	}
	if otherVisible[forumID] {
		t.Fatalf("otherVisible = %+v, want forum hidden", otherVisible)
	}
}

// TestVisibilityAuthorizerPermissionSettingsAndSimulation verifies admin grant persistence.
func TestVisibilityAuthorizerPermissionSettingsAndSimulation(t *testing.T) {
	_, _, db := newRepositories(t)
	store := orm.NewStore(db)
	authorizer := NewVisibilityAuthorizer(store)
	forumID := uuid.New()
	actorID := uuid.New()
	userID := uuid.New()
	groupID := uuid.New()
	memberID := uuid.New()
	createUser(t, db, actorID)
	createUser(t, db, userID)
	createUser(t, db, memberID)
	createGroup(t, db, groupID)
	createMembership(t, db, groupID, memberID)
	settings := domain.ForumPermissionSettings{
		ForumID:    forumID,
		Viewers:    []domain.ForumPermissionGrant{{SubjectType: domain.PermissionSubjectPublic}},
		Creators:   []domain.ForumPermissionGrant{{SubjectType: domain.PermissionSubjectUser, SubjectID: userID}},
		Moderators: []domain.ForumPermissionGrant{{SubjectType: domain.PermissionSubjectGroup, SubjectID: groupID}},
	}

	if err := authorizer.UpdateForumPermissionSettings(context.Background(), actorID, settings); err != nil {
		t.Fatalf("UpdateForumPermissionSettings() error = %v", err)
	}
	found, err := authorizer.ForumPermissionSettings(context.Background(), forumID)
	if err != nil {
		t.Fatalf("ForumPermissionSettings() error = %v", err)
	}
	if len(found.Viewers) != 1 || found.Viewers[0].SubjectID != domain.PublicPermissionSubjectID() || len(found.Creators) != 1 ||
		len(found.Moderators) != 1 {
		t.Fatalf("found = %+v, want persisted normalized grants", found)
	}
	publicResult, err := authorizer.SimulateForumPermission(
		context.Background(),
		forumID,
		domain.ForumPermissionSimulationRequest{Permission: string(groupsdomain.PermissionForumsView)},
	)
	if err != nil {
		t.Fatalf("Simulate public error = %v", err)
	}
	userResult, err := authorizer.SimulateForumPermission(
		context.Background(),
		forumID,
		domain.ForumPermissionSimulationRequest{ActorUserID: userID, Permission: string(groupsdomain.PermissionForumsCreateThread)},
	)
	if err != nil {
		t.Fatalf("Simulate user error = %v", err)
	}
	groupResult, err := authorizer.SimulateForumPermission(
		context.Background(),
		forumID,
		domain.ForumPermissionSimulationRequest{ActorUserID: memberID, Permission: string(groupsdomain.PermissionForumsManageThreads)},
	)
	if err != nil {
		t.Fatalf("Simulate group error = %v", err)
	}
	if !publicResult.Allowed || publicResult.MatchedRelation != string(groupsdomain.RelationViewer) || !userResult.Allowed ||
		userResult.MatchedRelation != string(groupsdomain.RelationCreator) ||
		!groupResult.Allowed ||
		groupResult.MatchedRelation != string(groupsdomain.RelationModerator) {
		t.Fatalf("results public=%+v user=%+v group=%+v, want matching explanations", publicResult, userResult, groupResult)
	}
	if err := authorizer.UpdateForumPermissionSettings(context.Background(), actorID, domain.ForumPermissionSettings{ForumID: forumID}); err != nil {
		t.Fatalf("UpdateForumPermissionSettings clear error = %v", err)
	}
	cleared, err := authorizer.ForumPermissionSettings(context.Background(), forumID)
	if err != nil {
		t.Fatalf("ForumPermissionSettings cleared error = %v", err)
	}
	if len(cleared.Viewers) != 0 || len(cleared.Creators) != 0 || len(cleared.Moderators) != 0 {
		t.Fatalf("cleared = %+v, want no managed grants", cleared)
	}
}

// TestVisibilityAuthorizerRejectsMissingGrantSubjects verifies user and group validation.
func TestVisibilityAuthorizerRejectsMissingGrantSubjects(t *testing.T) {
	_, _, db := newRepositories(t)
	authorizer := NewVisibilityAuthorizer(orm.NewStore(db))

	err := authorizer.UpdateForumPermissionSettings(
		context.Background(),
		uuid.New(),
		domain.ForumPermissionSettings{
			ForumID:  uuid.New(),
			Creators: []domain.ForumPermissionGrant{{SubjectType: domain.PermissionSubjectUser, SubjectID: uuid.New()}},
		},
	)
	var validation domain.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("UpdateForumPermissionSettings() error = %v, want validation", err)
	}
}

// TestThreadAndPostRepositoriesUpdateCounters verifies content counters.
func TestThreadAndPostRepositoriesUpdateCounters(t *testing.T) {
	categories, forums, db := newRepositories(t)
	threads := NewThreadRepository(orm.NewStore(db))
	posts := NewPostRepository(orm.NewStore(db))
	category, err := categories.Create(context.Background(), testCategory())
	if err != nil {
		t.Fatalf("Create category error = %v", err)
	}
	forum, err := forums.Create(context.Background(), testForum(category.ID, nil, "news"))
	if err != nil {
		t.Fatalf("Create forum error = %v", err)
	}
	authorID := uuid.New()
	openerID := uuid.New()
	thread := testThread(forum.ID, authorID, openerID)
	createdThread, err := threads.Create(context.Background(), thread)
	if err != nil {
		t.Fatalf("Create thread error = %v", err)
	}
	_, err = posts.Create(context.Background(), testPost(createdThread.ID, forum.ID, authorID, openerID, 1), nil)
	if err != nil {
		t.Fatalf("Create opener error = %v", err)
	}
	reply, err := posts.Create(context.Background(), testPost(createdThread.ID, forum.ID, uuid.New(), uuid.New(), 2), nil)
	if err != nil {
		t.Fatalf("Create reply error = %v", err)
	}

	foundThread, err := threads.FindByID(context.Background(), createdThread.ID)
	if err != nil {
		t.Fatalf("Find thread error = %v", err)
	}
	if foundThread.ReplyCount != 1 || foundThread.PostCount != 2 || foundThread.LatestPostID != reply.ID {
		t.Fatalf("thread counters = %+v, want reply/post/latest updates", foundThread)
	}
	stats, err := forums.ListStats(context.Background(), []uuid.UUID{forum.ID})
	if err != nil {
		t.Fatalf("ListStats error = %v", err)
	}
	if stats[forum.ID].ThreadCount != 1 || stats[forum.ID].PostCount != 2 {
		t.Fatalf("stats = %+v, want one thread and two posts", stats[forum.ID])
	}
}

// TestPostRepositoryUpdateWithRevisionPersistsHistory verifies revisions.
func TestPostRepositoryUpdateWithRevisionPersistsHistory(t *testing.T) {
	categories, forums, db := newRepositories(t)
	threads := NewThreadRepository(orm.NewStore(db))
	posts := NewPostRepository(orm.NewStore(db))
	category, err := categories.Create(context.Background(), testCategory())
	if err != nil {
		t.Fatalf("Create category error = %v", err)
	}
	forum, err := forums.Create(context.Background(), testForum(category.ID, nil, "support"))
	if err != nil {
		t.Fatalf("Create forum error = %v", err)
	}
	authorID := uuid.New()
	postID := uuid.New()
	thread, err := threads.Create(context.Background(), testThread(forum.ID, authorID, postID))
	if err != nil {
		t.Fatalf("Create thread error = %v", err)
	}
	post, err := posts.Create(context.Background(), testPost(thread.ID, forum.ID, authorID, postID, 1), nil)
	if err != nil {
		t.Fatalf("Create post error = %v", err)
	}
	updated := post
	updated.ContentDocumentJSON = []byte(`{"type":"doc","content":[]}`)
	updated.ContentText = "Edited"
	updated.EditCount = 1
	revision := domain.PostRevision{
		ID:                          uuid.New(),
		PostID:                      post.ID,
		EditedByUserID:              authorID,
		PreviousContentDocumentJSON: post.ContentDocumentJSON,
		PreviousContentText:         post.ContentText,
		EditReason:                  "typo",
	}

	if _, err := posts.UpdateWithRevision(context.Background(), updated, revision, post.Version); err != nil {
		t.Fatalf("UpdateWithRevision error = %v", err)
	}
	revisions, err := posts.ListRevisions(context.Background(), post.ID, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("ListRevisions error = %v", err)
	}
	if len(revisions.Items) != 1 || revisions.Items[0].PreviousContentText != post.ContentText {
		t.Fatalf("revisions = %+v, want previous content", revisions.Items)
	}
}

// TestPostRepositoryStoresReferencesAndNextSequence verifies references and sequencing.
func TestPostRepositoryStoresReferencesAndNextSequence(t *testing.T) {
	categories, forums, db := newRepositories(t)
	threads := NewThreadRepository(orm.NewStore(db))
	posts := NewPostRepository(orm.NewStore(db))
	category, err := categories.Create(context.Background(), testCategory())
	if err != nil {
		t.Fatalf("Create category error = %v", err)
	}
	forum, err := forums.Create(context.Background(), testForum(category.ID, nil, "refs"))
	if err != nil {
		t.Fatalf("Create forum error = %v", err)
	}
	authorID := uuid.New()
	openerID := uuid.New()
	thread, err := threads.Create(context.Background(), testThread(forum.ID, authorID, openerID))
	if err != nil {
		t.Fatalf("Create thread error = %v", err)
	}
	opener, err := posts.Create(context.Background(), testPost(thread.ID, forum.ID, authorID, openerID, 1), nil)
	if err != nil {
		t.Fatalf("Create opener error = %v", err)
	}
	replyID := uuid.New()
	reference := domain.PostReference{
		ID:            uuid.New(),
		SourcePostID:  replyID,
		TargetPostID:  &opener.ID,
		ReferenceType: domain.ReferenceQuote,
		QuoteExcerpt:  "Original",
	}
	reply, err := posts.Create(
		context.Background(),
		testPost(thread.ID, forum.ID, uuid.New(), replyID, 2),
		[]domain.PostReference{reference},
	)
	if err != nil {
		t.Fatalf("Create reply error = %v", err)
	}
	next, err := posts.NextSequence(context.Background(), thread.ID)
	if err != nil {
		t.Fatalf("NextSequence error = %v", err)
	}
	refs, err := posts.ListReferences(context.Background(), []uuid.UUID{reply.ID})
	if err != nil {
		t.Fatalf("ListReferences error = %v", err)
	}
	if next != 3 || len(refs[reply.ID]) != 1 || refs[reply.ID][0].QuoteExcerpt != "Original" {
		t.Fatalf("next=%d refs=%+v, want next sequence and quote ref", next, refs)
	}
}

// TestInteractionRepositoryLikeUnlikeIdempotency verifies like counter safety.
func TestInteractionRepositoryLikeUnlikeIdempotency(t *testing.T) {
	categories, forums, db := newRepositories(t)
	threads := NewThreadRepository(orm.NewStore(db))
	posts := NewPostRepository(orm.NewStore(db))
	interactions := NewInteractionRepository(orm.NewStore(db))
	_, forum, thread, post := createContentFixture(t, categories, forums, threads, posts, "likes")
	userID := uuid.New()
	like := domain.PostLike{
		ID:        uuid.New(),
		PostID:    post.ID,
		ThreadID:  thread.ID,
		ForumID:   forum.ID,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
	}

	changed, err := interactions.LikePost(context.Background(), like)
	if err != nil {
		t.Fatalf("LikePost first error = %v", err)
	}
	again, err := interactions.LikePost(context.Background(), like)
	if err != nil {
		t.Fatalf("LikePost second error = %v", err)
	}
	if !changed || again {
		t.Fatalf("changed=%v again=%v, want first change only", changed, again)
	}
	liked, err := interactions.LikedByUser(context.Background(), post.ID, userID)
	if err != nil {
		t.Fatalf("LikedByUser error = %v", err)
	}
	if !liked {
		t.Fatalf("LikedByUser = false, want true")
	}
	removed, err := interactions.UnlikePost(context.Background(), post.ID, userID)
	if err != nil {
		t.Fatalf("UnlikePost first error = %v", err)
	}
	missing, err := interactions.UnlikePost(context.Background(), post.ID, userID)
	if err != nil {
		t.Fatalf("UnlikePost second error = %v", err)
	}
	if !removed || missing {
		t.Fatalf("removed=%v missing=%v, want first unlike only", removed, missing)
	}
	found, err := posts.FindByID(context.Background(), post.ID)
	if err != nil {
		t.Fatalf("FindByID error = %v", err)
	}
	if found.LikeCount != 0 {
		t.Fatalf("post like count = %d, want 0", found.LikeCount)
	}
}

// TestInteractionRepositoryWidgetsReturnVisibleRows verifies widget queries.
func TestInteractionRepositoryWidgetsReturnVisibleRows(t *testing.T) {
	categories, forums, db := newRepositories(t)
	threads := NewThreadRepository(orm.NewStore(db))
	posts := NewPostRepository(orm.NewStore(db))
	interactions := NewInteractionRepository(orm.NewStore(db))
	_, forum, thread, post := createContentFixture(t, categories, forums, threads, posts, "widgets")
	if _, err := interactions.LikePost(context.Background(), domain.PostLike{ID: uuid.New(), PostID: post.ID, ThreadID: thread.ID, ForumID: forum.ID, UserID: uuid.New(), CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("LikePost error = %v", err)
	}

	latest, err := interactions.ListLatestPosts(
		context.Background(),
		port.LatestPostFilter{ForumIDs: []uuid.UUID{forum.ID}},
		pagination.Page{Limit: 10},
	)
	if err != nil {
		t.Fatalf("ListLatestPosts error = %v", err)
	}
	mostLiked, err := interactions.ListMostLikedPosts(
		context.Background(),
		port.MostLikedFilter{ForumID: forum.ID},
		pagination.Page{Limit: 10},
	)
	if err != nil {
		t.Fatalf("ListMostLikedPosts error = %v", err)
	}
	if len(latest.Items) != 1 || latest.Items[0].PostID != post.ID {
		t.Fatalf("latest = %+v, want post", latest.Items)
	}
	if len(mostLiked.Items) != 1 || mostLiked.Items[0].LikeCount != 1 {
		t.Fatalf("mostLiked = %+v, want liked post", mostLiked.Items)
	}
}

// TestInteractionRepositoryReadStateAndUnreadSummary verifies read-state persistence.
func TestInteractionRepositoryReadStateAndUnreadSummary(t *testing.T) {
	categories, forums, db := newRepositories(t)
	threads := NewThreadRepository(orm.NewStore(db))
	posts := NewPostRepository(orm.NewStore(db))
	interactions := NewInteractionRepository(orm.NewStore(db))
	_, forum, thread, _ := createContentFixture(t, categories, forums, threads, posts, "reads")
	userID := uuid.New()

	before, err := interactions.UnreadSummary(context.Background(), userID, []uuid.UUID{forum.ID})
	if err != nil {
		t.Fatalf("UnreadSummary before error = %v", err)
	}
	if before.UnreadThreadCount != 1 {
		t.Fatalf("before unread = %+v, want one unread thread", before)
	}
	state := domain.ThreadReadState{
		ID:                   uuid.New(),
		UserID:               userID,
		ForumID:              forum.ID,
		ThreadID:             thread.ID,
		LastReadPostSequence: 1,
		LastReadAt:           time.Now().UTC(),
	}
	if err := interactions.MarkThreadRead(context.Background(), state); err != nil {
		t.Fatalf("MarkThreadRead error = %v", err)
	}
	after, err := interactions.UnreadSummary(context.Background(), userID, []uuid.UUID{forum.ID})
	if err != nil {
		t.Fatalf("UnreadSummary after error = %v", err)
	}
	if after.UnreadThreadCount != 0 {
		t.Fatalf("after unread = %+v, want zero unread threads", after)
	}
	if err := interactions.MarkForumRead(context.Background(), userID, forum.ID, time.Now().UTC()); err != nil {
		t.Fatalf("MarkForumRead error = %v", err)
	}
}

// TestOperationsRepositorySearchReturnsThreadsAndPosts verifies forum search reads source tables.
func TestOperationsRepositorySearchReturnsThreadsAndPosts(t *testing.T) {
	categories, forums, db := newRepositories(t)
	threads := NewThreadRepository(orm.NewStore(db))
	posts := NewPostRepository(orm.NewStore(db))
	operations := NewOperationsRepository(orm.NewStore(db))
	_, forum, thread, post := createContentFixture(t, categories, forums, threads, posts, "search")
	thread.Title = "Alpha launch notes"
	if _, err := threads.UpdateTitle(context.Background(), thread, thread.Version); err != nil {
		t.Fatalf("UpdateTitle error = %v", err)
	}
	post.ContentText = "Alpha reply body"
	if _, err := posts.UpdateWithRevision(context.Background(), post, domain.PostRevision{ID: uuid.New(), PostID: post.ID, EditedByUserID: post.AuthorUserID, PreviousContentDocumentJSON: post.ContentDocumentJSON, PreviousContentText: "Original"}, post.Version); err != nil {
		t.Fatalf("UpdateWithRevision error = %v", err)
	}

	result, err := operations.Search(
		context.Background(),
		port.SearchFilter{ForumIDs: []uuid.UUID{forum.ID}, Query: "alpha"},
		pagination.Page{Limit: 10},
	)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("items = %+v, want thread and post match", result.Items)
	}
}

// TestOperationsRepositoryStatsVerifyAndRebuild verifies counter reconciliation.
func TestOperationsRepositoryStatsVerifyAndRebuild(t *testing.T) {
	categories, forums, db := newRepositories(t)
	threads := NewThreadRepository(orm.NewStore(db))
	posts := NewPostRepository(orm.NewStore(db))
	operations := NewOperationsRepository(orm.NewStore(db))
	_, forum, thread, _ := createContentFixture(t, categories, forums, threads, posts, "stats")
	if err := db.Model(&ThreadModel{}).Where("id = ?", thread.ID).Updates(map[string]any{"post_count": 99, "visible_post_count": 99, "reply_count": 99, "visible_reply_count": 99}).Error; err != nil {
		t.Fatalf("corrupt thread counters error = %v", err)
	}
	if err := db.Model(&StatsModel{}).Where("forum_id = ?", forum.ID).Updates(map[string]any{"thread_count": 99, "visible_thread_count": 99, "post_count": 99, "visible_post_count": 99}).Error; err != nil {
		t.Fatalf("corrupt forum stats error = %v", err)
	}

	drift, err := operations.VerifyStats(context.Background())
	if err != nil {
		t.Fatalf("VerifyStats() error = %v", err)
	}
	if len(drift.Mismatches) == 0 {
		t.Fatalf("VerifyStats() mismatches = empty, want drift")
	}
	rebuilt, err := operations.RebuildStats(context.Background())
	if err != nil {
		t.Fatalf("RebuildStats() error = %v", err)
	}
	if !rebuilt.Repaired {
		t.Fatalf("rebuilt = %+v, want repaired report", rebuilt)
	}
	clean, err := operations.VerifyStats(context.Background())
	if err != nil {
		t.Fatalf("VerifyStats clean error = %v", err)
	}
	if len(clean.Mismatches) != 0 {
		t.Fatalf("clean mismatches = %+v, want none", clean.Mismatches)
	}
}

// TestOperationsRepositoryLikesVerifyRebuildAndViews verifies like and view reconciliation.
func TestOperationsRepositoryLikesVerifyRebuildAndViews(t *testing.T) {
	categories, forums, db := newRepositories(t)
	threads := NewThreadRepository(orm.NewStore(db))
	posts := NewPostRepository(orm.NewStore(db))
	interactions := NewInteractionRepository(orm.NewStore(db))
	operations := NewOperationsRepository(orm.NewStore(db))
	_, _, thread, post := createContentFixture(t, categories, forums, threads, posts, "ops")
	if _, err := interactions.LikePost(context.Background(), domain.PostLike{ID: uuid.New(), PostID: post.ID, ThreadID: thread.ID, ForumID: post.ForumID, UserID: uuid.New(), CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("LikePost error = %v", err)
	}
	if err := db.Model(&PostModel{}).Where("id = ?", post.ID).Update("like_count", 0).Error; err != nil {
		t.Fatalf("corrupt post like error = %v", err)
	}
	if err := db.Model(&ThreadModel{}).Where("id = ?", thread.ID).Updates(map[string]any{"like_count": 0, "view_count": 0}).Error; err != nil {
		t.Fatalf("corrupt thread like error = %v", err)
	}

	drift, err := operations.VerifyLikes(context.Background())
	if err != nil {
		t.Fatalf("VerifyLikes() error = %v", err)
	}
	if len(drift.Mismatches) != 2 {
		t.Fatalf("like drift = %+v, want post and thread mismatches", drift.Mismatches)
	}
	if _, err := operations.RebuildLikes(context.Background()); err != nil {
		t.Fatalf("RebuildLikes() error = %v", err)
	}
	if err := operations.ApplyThreadViews(context.Background(), map[uuid.UUID]int64{thread.ID: 4}); err != nil {
		t.Fatalf("ApplyThreadViews() error = %v", err)
	}
	found, err := threads.FindByID(context.Background(), thread.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.LikeCount != 1 || found.ViewCount != 4 {
		t.Fatalf("thread counters = %+v, want repaired like and flushed views", found)
	}
}

// newRepositories creates migrated forum repositories.
func newRepositories(t *testing.T) (CategoryRepository, ForumRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("migrate Up() error = %v", err)
	}
	store := orm.NewStore(db)
	return NewCategoryRepository(store), NewForumRepository(store), db
}

// createContentFixture creates one category, forum, thread, and opener post.
func createContentFixture(
	t *testing.T,
	categories CategoryRepository,
	forums ForumRepository,
	threads ThreadRepository,
	posts PostRepository,
	key string,
) (domain.ForumCategory, domain.Forum, domain.Thread, domain.Post) {
	t.Helper()
	category, err := categories.Create(context.Background(), testCategory())
	if err != nil {
		t.Fatalf("Create category error = %v", err)
	}
	forum, err := forums.Create(context.Background(), testForum(category.ID, nil, key))
	if err != nil {
		t.Fatalf("Create forum error = %v", err)
	}
	authorID := uuid.New()
	openerID := uuid.New()
	thread, err := threads.Create(context.Background(), testThread(forum.ID, authorID, openerID))
	if err != nil {
		t.Fatalf("Create thread error = %v", err)
	}
	post, err := posts.Create(context.Background(), testPost(thread.ID, forum.ID, authorID, openerID, 1), nil)
	if err != nil {
		t.Fatalf("Create opener error = %v", err)
	}
	return category, forum, thread, post
}

// createTuple stores one visibility tuple.
func createTuple(t *testing.T, db *gorm.DB, forumID uuid.UUID, subjectType groupsdomain.SubjectType, subjectID uuid.UUID) {
	t.Helper()
	subjectRelation := ""
	if subjectType == groupsdomain.SubjectGroup {
		subjectRelation = string(groupsdomain.RelationMember)
	}
	err := db.Exec(
		"INSERT INTO authorization_relation_tuples (id, object_type, object_id, relation, subject_type, subject_id, subject_relation, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)",
		uuid.New(),
		groupsdomain.ObjectForum,
		forumID,
		groupsdomain.RelationViewer,
		subjectType,
		subjectID,
		subjectRelation,
	).Error
	if err != nil {
		t.Fatalf("insert tuple error = %v", err)
	}
}

// createGroup stores one active group.
func createGroup(t *testing.T, db *gorm.DB, groupID uuid.UUID) {
	t.Helper()
	err := db.Exec(
		"INSERT INTO groups (id, key, name, description, color, weight, status, version, created_at, updated_at) VALUES (?, ?, ?, '', '#ffffff', 0, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		groupID,
		"mods"+groupID.String()[:8],
		"Moderators",
		groupsdomain.GroupStatusActive,
	).Error
	if err != nil {
		t.Fatalf("insert group error = %v", err)
	}
}

// createMembership stores one active group membership.
func createMembership(t *testing.T, db *gorm.DB, groupID uuid.UUID, userID uuid.UUID) {
	t.Helper()
	err := db.Exec(
		"INSERT INTO group_memberships (id, group_id, user_id, status, version, created_at, updated_at) VALUES (?, ?, ?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		uuid.New(),
		groupID,
		userID,
		groupsdomain.MembershipStatusActive,
	).Error
	if err != nil {
		t.Fatalf("insert membership error = %v", err)
	}
}

// createUser stores one active user.
func createUser(t *testing.T, db *gorm.DB, userID uuid.UUID) {
	t.Helper()
	err := db.Exec(
		"INSERT INTO users (id, status, first_seen_at, version, created_at, updated_at) VALUES (?, 'active', CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		userID,
	).Error
	if err != nil {
		t.Fatalf("insert user error = %v", err)
	}
}

// testCategory returns a persisted category.
func testCategory() domain.ForumCategory {
	return domain.ForumCategory{ID: uuid.New(), Key: "official", Name: "Official", Status: domain.CategoryStatusActive, Version: 1}
}

// testForum returns a persisted forum.
func testForum(categoryID uuid.UUID, parentID *uuid.UUID, key string) domain.Forum {
	id := uuid.New()
	return domain.Forum{
		ID:                            id,
		CategoryID:                    categoryID,
		ParentForumID:                 parentID,
		Kind:                          domain.ForumKindDiscussion,
		Key:                           domain.Key(key),
		Slug:                          domain.Slug(key),
		Name:                          key,
		Path:                          "/" + id.String() + "/",
		ThreadVisibilityMode:          domain.ThreadVisibilityAllThreads,
		DefaultThreadStatus:           domain.ThreadStatusOpen,
		AuthorPostEditWindowSeconds:   domain.DefaultAuthorPostEditWindowSeconds,
		AuthorPostDeleteWindowSeconds: domain.DefaultAuthorPostDeleteWindowSeconds,
		Status:                        domain.ForumStatusActive,
		Version:                       1,
	}
}

// testThread returns a persisted thread.
func testThread(forumID uuid.UUID, authorID uuid.UUID, openerID uuid.UUID) domain.Thread {
	return domain.Thread{
		ID:                     uuid.New(),
		ForumID:                forumID,
		AuthorUserID:           authorID,
		OpenerPostID:           openerID,
		LatestPostID:           openerID,
		LatestPostAuthorUserID: authorID,
		LatestPostAt:           time.Now().UTC(),
		Title:                  "A thread",
		Slug:                   "a-thread",
		Status:                 domain.ThreadStatusOpen,
		StickyState:            domain.StickyStateNormal,
		PostCount:              1,
		VisiblePostCount:       1,
		Version:                1,
	}
}

// testPost returns a persisted post.
func testPost(threadID uuid.UUID, forumID uuid.UUID, authorID uuid.UUID, postID uuid.UUID, sequence int64) domain.Post {
	return domain.Post{
		ID:                  postID,
		ThreadID:            threadID,
		ForumID:             forumID,
		AuthorUserID:        authorID,
		Sequence:            sequence,
		Status:              domain.PostStatusVisible,
		ContentFormat:       domain.ContentFormatProseMirror,
		ContentDocumentJSON: []byte(`{"type":"doc"}`),
		ContentText:         "Original",
		Version:             1,
	}
}
