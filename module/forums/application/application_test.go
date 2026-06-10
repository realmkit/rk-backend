package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// TestServiceCreateForumBuildsPathAndStats verifies forum creation prepares tree fields.
func TestServiceCreateForumBuildsPathAndStats(t *testing.T) {
	service, categories, forums, auth, _ := newTestService()
	actorID := uuid.New()
	auth.manage[domain.RootForumObjectID()] = true
	category := testCategory()
	categories.items[category.ID] = category

	created, err := service.CreateForum(context.Background(), port.CreateForumCommand{ActorUserID: actorID, Forum: domain.Forum{CategoryID: category.ID, Key: "news", Slug: "news", Name: "News"}})
	if err != nil {
		t.Fatalf("CreateForum() error = %v", err)
	}
	if created.Path != "/"+created.ID.String()+"/" || created.Depth != 0 {
		t.Fatalf("created path=%q depth=%d, want root path", created.Path, created.Depth)
	}
	if _, ok := forums.stats[created.ID]; !ok {
		t.Fatalf("stats missing for created forum")
	}
}

// TestServiceCreateCategoryRequiresManagePermission verifies structure admin checks.
func TestServiceCreateCategoryRequiresManagePermission(t *testing.T) {
	service, _, _, _, _ := newTestService()

	_, err := service.CreateCategory(context.Background(), port.CreateCategoryCommand{ActorUserID: uuid.New(), Category: testCategory()})
	if !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("CreateCategory() error = %v, want %v", err, port.ErrForbidden)
	}
}

// TestServiceCreateForumUnderParentInheritsParentCategory verifies tree preparation.
func TestServiceCreateForumUnderParentInheritsParentCategory(t *testing.T) {
	service, categories, forums, auth, _ := newTestService()
	actorID := uuid.New()
	auth.manage[domain.RootForumObjectID()] = true
	rootCategory := testCategory()
	otherCategory := testCategory()
	otherCategory.ID = uuid.New()
	otherCategory.Key = "other"
	categories.items[rootCategory.ID] = rootCategory
	categories.items[otherCategory.ID] = otherCategory
	parent := testForum(rootCategory.ID, nil, 0, "support")
	forums.items[parent.ID] = parent

	created, err := service.CreateForum(context.Background(), port.CreateForumCommand{ActorUserID: actorID, Forum: domain.Forum{CategoryID: otherCategory.ID, ParentForumID: &parent.ID, Key: "help", Slug: "help", Name: "Help"}})
	if err != nil {
		t.Fatalf("CreateForum() error = %v", err)
	}
	if created.CategoryID != rootCategory.ID || created.Depth != 1 || created.Path != parent.Path+created.ID.String()+"/" {
		t.Fatalf("created = %+v, want child under parent category/path", created)
	}
}

// TestServiceCreateForumRejectsLinkParent verifies links cannot contain children.
func TestServiceCreateForumRejectsLinkParent(t *testing.T) {
	service, categories, forums, auth, _ := newTestService()
	actorID := uuid.New()
	auth.manage[domain.RootForumObjectID()] = true
	category := testCategory()
	categories.items[category.ID] = category
	parent := testForum(category.ID, nil, 0, "discord")
	parent.Kind = domain.ForumKindLink
	parent.ExternalURL = "https://discord.example"
	forums.items[parent.ID] = parent

	_, err := service.CreateForum(context.Background(), port.CreateForumCommand{ActorUserID: actorID, Forum: domain.Forum{CategoryID: category.ID, ParentForumID: &parent.ID, Key: "child", Slug: "child", Name: "Child"}})
	if !errors.Is(err, port.ErrInvalidMove) {
		t.Fatalf("CreateForum() error = %v, want %v", err, port.ErrInvalidMove)
	}
}

// TestServiceTreeFiltersInvisibleForumsAndCaches verifies tree visibility and cache.
func TestServiceTreeFiltersInvisibleForumsAndCaches(t *testing.T) {
	service, categories, forums, auth, cache := newTestService()
	category := testCategory()
	categories.items[category.ID] = category
	visible := testForum(category.ID, nil, 0, "visible")
	hidden := testForum(category.ID, nil, 1, "hidden")
	forums.items[visible.ID] = visible
	forums.items[hidden.ID] = hidden
	forums.stats[visible.ID] = domain.ForumStats{ForumID: visible.ID}
	forums.stats[hidden.ID] = domain.ForumStats{ForumID: hidden.ID}
	auth.visible[visible.ID] = true

	tree, err := service.Tree(context.Background(), uuid.Nil)
	if err != nil {
		t.Fatalf("Tree() error = %v", err)
	}
	if len(tree.Categories) != 1 || len(tree.Categories[0].Forums) != 1 || tree.Categories[0].Forums[0].Forum.ID != visible.ID {
		t.Fatalf("tree = %+v, want only visible forum", tree)
	}
	forums.items = map[uuid.UUID]domain.Forum{}
	cached, err := service.Tree(context.Background(), uuid.Nil)
	if err != nil {
		t.Fatalf("cached Tree() error = %v", err)
	}
	if len(cached.Categories) != 1 || cache.sets != 1 {
		t.Fatalf("cached tree=%+v sets=%d, want cache hit after one set", cached, cache.sets)
	}
}

// TestServiceReorderCategoriesRejectsInvalidItems verifies reorder validation.
func TestServiceReorderCategoriesRejectsInvalidItems(t *testing.T) {
	service, _, _, auth, _ := newTestService()
	auth.manage[domain.RootForumObjectID()] = true

	err := service.ReorderCategories(context.Background(), port.ReorderCategoriesCommand{ActorUserID: uuid.New()})
	var validation domain.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("ReorderCategories() error = %v, want validation error", err)
	}
}

// TestServiceMoveForumRejectsDescendantParent verifies invalid tree moves.
func TestServiceMoveForumRejectsDescendantParent(t *testing.T) {
	service, categories, forums, auth, _ := newTestService()
	actorID := uuid.New()
	category := testCategory()
	categories.items[category.ID] = category
	parent := testForum(category.ID, nil, 0, "parent")
	child := testForum(category.ID, &parent.ID, 1, "child")
	child.Path = parent.Path + child.ID.String() + "/"
	forums.items[parent.ID] = parent
	forums.items[child.ID] = child
	auth.manage[parent.ID] = true

	_, err := service.MoveForum(context.Background(), port.MoveForumCommand{ActorUserID: actorID, ID: parent.ID, CategoryID: category.ID, ParentForumID: &child.ID, ExpectedVersion: parent.Version})
	if !errors.Is(err, port.ErrInvalidMove) {
		t.Fatalf("MoveForum() error = %v, want %v", err, port.ErrInvalidMove)
	}
}

// TestServiceCreateThreadCreatesOpenerPost verifies thread creation transaction inputs.
func TestServiceCreateThreadCreatesOpenerPost(t *testing.T) {
	service, categories, forums, threads, posts, auth, _ := newContentTestService()
	actorID := uuid.New()
	category := testCategory()
	forum := testForum(category.ID, nil, 0, "news")
	categories.items[category.ID] = category
	forums.items[forum.ID] = forum
	auth.create[forum.ID] = true

	thread, post, err := service.CreateThread(context.Background(), port.CreateThreadCommand{ActorUserID: actorID, ForumID: forum.ID, Title: "My first thread", Slug: "my-first-thread", ContentDocumentJSON: []byte(`{"type":"doc","content":[{"type":"text","text":"Hello from JSON"}]}`)})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	if thread.OpenerPostID != post.ID || post.Sequence != 1 || post.ContentText != "Hello from JSON" || threads.items[thread.ID].ID != thread.ID || posts.items[post.ID].ID != post.ID {
		t.Fatalf("thread=%+v post=%+v, want opener post stored", thread, post)
	}
}

// TestServiceCreateReplyRequiresOpenThread verifies closed reply state rules.
func TestServiceCreateReplyRequiresOpenThread(t *testing.T) {
	service, categories, forums, threads, _, auth, _ := newContentTestService()
	actorID := uuid.New()
	category := testCategory()
	forum := testForum(category.ID, nil, 0, "support")
	thread := testThread(forum.ID, actorID)
	thread.Status = domain.ThreadStatusClosed
	categories.items[category.ID] = category
	forums.items[forum.ID] = forum
	threads.items[thread.ID] = thread
	auth.reply[forum.ID] = true

	_, err := service.CreateReply(context.Background(), port.CreateReplyCommand{ActorUserID: actorID, ThreadID: thread.ID, ContentDocumentJSON: []byte(`{"type":"doc"}`), ContentText: "Reply"})
	if !errors.Is(err, port.ErrConflict) {
		t.Fatalf("CreateReply() error = %v, want %v", err, port.ErrConflict)
	}
}

// TestServiceCreateReplyRejectsLockedThread verifies locked threads cannot receive replies.
func TestServiceCreateReplyRejectsLockedThread(t *testing.T) {
	service, categories, forums, threads, _, auth, _ := newContentTestService()
	actorID := uuid.New()
	category := testCategory()
	forum := testForum(category.ID, nil, 0, "locked")
	thread := testThread(forum.ID, actorID)
	thread.Status = domain.ThreadStatusLocked
	categories.items[category.ID] = category
	forums.items[forum.ID] = forum
	threads.items[thread.ID] = thread
	auth.reply[forum.ID] = true

	_, err := service.CreateReply(context.Background(), port.CreateReplyCommand{ActorUserID: actorID, ThreadID: thread.ID, ContentDocumentJSON: []byte(`{"type":"doc"}`), ContentText: "Reply"})
	if !errors.Is(err, port.ErrConflict) {
		t.Fatalf("CreateReply() error = %v, want %v", err, port.ErrConflict)
	}
}

// TestServiceCreateReplyStoresNextSequence verifies reply sequence allocation.
func TestServiceCreateReplyStoresNextSequence(t *testing.T) {
	service, _, forums, threads, posts, auth, _ := newContentTestService()
	actorID := uuid.New()
	forum := testForum(uuid.New(), nil, 0, "games")
	thread := testThread(forum.ID, uuid.New())
	forums.items[forum.ID] = forum
	threads.items[thread.ID] = thread
	posts.items[thread.OpenerPostID] = testPost(thread.ID, forum.ID, thread.AuthorUserID, 1)
	auth.reply[forum.ID] = true

	reply, err := service.CreateReply(context.Background(), port.CreateReplyCommand{ActorUserID: actorID, ThreadID: thread.ID, ContentDocumentJSON: []byte(`{"type":"doc"}`), ContentText: "Nice"})
	if err != nil {
		t.Fatalf("CreateReply() error = %v", err)
	}
	if reply.Sequence != 2 || posts.items[reply.ID].ID != reply.ID {
		t.Fatalf("reply = %+v, want sequence 2 stored reply", reply)
	}
}

// TestServiceCreateReplyRejectsMissingAttachment verifies asset references are validated.
func TestServiceCreateReplyRejectsMissingAttachment(t *testing.T) {
	service, _, forums, threads, _, auth, _ := newContentTestService()
	actorID := uuid.New()
	forum := testForum(uuid.New(), nil, 0, "attachments")
	thread := testThread(forum.ID, actorID)
	assetID := uuid.New()
	forums.items[forum.ID] = forum
	threads.items[thread.ID] = thread
	auth.reply[forum.ID] = true
	auth.visible[forum.ID] = true

	_, err := service.CreateReply(context.Background(), port.CreateReplyCommand{ActorUserID: actorID, ThreadID: thread.ID, ContentDocumentJSON: []byte(`{"type":"doc"}`), ContentText: "Reply", References: []domain.PostReference{{TargetAssetID: &assetID, ReferenceType: domain.ReferenceAttachment}}})
	if !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("CreateReply() error = %v, want %v", err, port.ErrNotFound)
	}
}

// TestServiceUpdatePostWritesRevision verifies edit history is preserved.
func TestServiceUpdatePostWritesRevision(t *testing.T) {
	service, _, _, threads, posts, _, _ := newContentTestService()
	actorID := uuid.New()
	thread := testThread(uuid.New(), actorID)
	post := testPost(thread.ID, thread.ForumID, actorID, 1)
	threads.items[thread.ID] = thread
	posts.items[post.ID] = post

	updated, err := service.UpdatePost(context.Background(), port.UpdatePostCommand{ActorUserID: actorID, PostID: post.ID, ContentDocumentJSON: []byte(`{"type":"doc","content":[{"type":"paragraph"}]}`), ContentText: "Edited", EditReason: "typo", ExpectedVersion: post.Version})
	if err != nil {
		t.Fatalf("UpdatePost() error = %v", err)
	}
	if updated.ContentText != "Edited" || updated.EditCount != 1 || len(posts.revisions[post.ID]) != 1 || posts.revisions[post.ID][0].PreviousContentText != post.ContentText {
		t.Fatalf("updated=%+v revisions=%+v, want edited post and revision", updated, posts.revisions[post.ID])
	}
}

// TestServiceUpdatePostRejectsExpiredAuthorWindow verifies edit window policy.
func TestServiceUpdatePostRejectsExpiredAuthorWindow(t *testing.T) {
	service, _, _, threads, posts, _, _ := newContentTestService()
	actorID := uuid.New()
	thread := testThread(uuid.New(), actorID)
	post := testPost(thread.ID, thread.ForumID, actorID, 1)
	post.CreatedAt = time.Now().UTC().Add(-authorPostEditWindow - time.Minute)
	threads.items[thread.ID] = thread
	posts.items[post.ID] = post

	_, err := service.UpdatePost(context.Background(), port.UpdatePostCommand{ActorUserID: actorID, PostID: post.ID, ContentDocumentJSON: []byte(`{"type":"doc"}`), ContentText: "Too late", ExpectedVersion: post.Version})
	if !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("UpdatePost() error = %v, want %v", err, port.ErrForbidden)
	}
}

// TestServiceGetThreadAllowsVisibleForum verifies thread visibility through forum grants.
func TestServiceGetThreadAllowsVisibleForum(t *testing.T) {
	service, _, forums, threads, _, auth, _ := newContentTestService()
	actorID := uuid.New()
	forum := testForum(uuid.New(), nil, 0, "visible")
	thread := testThread(forum.ID, uuid.New())
	forums.items[forum.ID] = forum
	threads.items[thread.ID] = thread
	auth.visible[forum.ID] = true

	found, err := service.GetThread(context.Background(), actorID, thread.ID)
	if err != nil {
		t.Fatalf("GetThread() error = %v", err)
	}
	if found.ID != thread.ID {
		t.Fatalf("found = %+v, want thread", found)
	}
}

// TestServiceUpdateThreadTitleRequiresManageForNonAuthor verifies title edit gates.
func TestServiceUpdateThreadTitleRequiresManageForNonAuthor(t *testing.T) {
	service, _, forums, threads, _, auth, _ := newContentTestService()
	actorID := uuid.New()
	forum := testForum(uuid.New(), nil, 0, "threads")
	thread := testThread(forum.ID, uuid.New())
	forums.items[forum.ID] = forum
	threads.items[thread.ID] = thread

	_, err := service.UpdateThreadTitle(context.Background(), port.UpdateThreadTitleCommand{ActorUserID: actorID, ThreadID: thread.ID, Title: "Changed title", Slug: "changed-title", ExpectedVersion: thread.Version})
	if !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("UpdateThreadTitle() error = %v, want %v", err, port.ErrForbidden)
	}
	auth.manageThreads[forum.ID] = true
	if _, err := service.UpdateThreadTitle(context.Background(), port.UpdateThreadTitleCommand{ActorUserID: actorID, ThreadID: thread.ID, Title: "Changed title", Slug: "changed-title", ExpectedVersion: thread.Version}); err != nil {
		t.Fatalf("UpdateThreadTitle() allowed error = %v", err)
	}
}

// TestServiceListPostsIncludeHiddenRequiresManagePermission verifies hidden post gates.
func TestServiceListPostsIncludeHiddenRequiresManagePermission(t *testing.T) {
	service, _, forums, threads, posts, auth, _ := newContentTestService()
	actorID := uuid.New()
	forum := testForum(uuid.New(), nil, 0, "staff")
	thread := testThread(forum.ID, uuid.New())
	post := testPost(thread.ID, forum.ID, thread.AuthorUserID, 1)
	post.Status = domain.PostStatusHidden
	forums.items[forum.ID] = forum
	threads.items[thread.ID] = thread
	posts.items[post.ID] = post

	_, err := service.ListPosts(context.Background(), actorID, port.PostFilter{ThreadID: thread.ID, IncludeHidden: true}, pagination.Page{Limit: 10})
	if !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("ListPosts() error = %v, want %v", err, port.ErrForbidden)
	}
	auth.managePosts[forum.ID] = true
	result, err := service.ListPosts(context.Background(), actorID, port.PostFilter{ThreadID: thread.ID, IncludeHidden: true}, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("ListPosts() allowed error = %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items = %d, want hidden post", len(result.Items))
	}
}

// TestServiceDeletePostRequiresManageForNonAuthor verifies delete gates.
func TestServiceDeletePostRequiresManageForNonAuthor(t *testing.T) {
	service, _, _, _, posts, auth, _ := newContentTestService()
	actorID := uuid.New()
	post := testPost(uuid.New(), uuid.New(), uuid.New(), 1)
	posts.items[post.ID] = post

	err := service.DeletePost(context.Background(), port.DeletePostCommand{ActorUserID: actorID, PostID: post.ID, ExpectedVersion: post.Version})
	if !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("DeletePost() error = %v, want %v", err, port.ErrForbidden)
	}
	auth.managePosts[post.ForumID] = true
	if err := service.DeletePost(context.Background(), port.DeletePostCommand{ActorUserID: actorID, PostID: post.ID, ExpectedVersion: post.Version}); err != nil {
		t.Fatalf("DeletePost() allowed error = %v", err)
	}
}

// TestServiceListPostRevisionsRequiresManagePermission verifies revision gates.
func TestServiceListPostRevisionsRequiresManagePermission(t *testing.T) {
	service, _, forums, _, posts, auth, _ := newContentTestService()
	actorID := uuid.New()
	post := testPost(uuid.New(), uuid.New(), uuid.New(), 1)
	forums.items[post.ForumID] = testForum(uuid.New(), nil, 0, "mods")
	posts.items[post.ID] = post

	_, err := service.ListPostRevisions(context.Background(), actorID, post.ID, pagination.Page{Limit: 10})
	if !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("ListPostRevisions() error = %v, want %v", err, port.ErrForbidden)
	}
	auth.managePosts[post.ForumID] = true
	if _, err := service.ListPostRevisions(context.Background(), actorID, post.ID, pagination.Page{Limit: 10}); err != nil {
		t.Fatalf("ListPostRevisions() allowed error = %v", err)
	}
}

// TestServiceLikePostIsIdempotent verifies like commands do not drift counts.
func TestServiceLikePostIsIdempotent(t *testing.T) {
	service, _, forums, threads, posts, auth, interactions := newContentTestService()
	actorID := uuid.New()
	forum := testForum(uuid.New(), nil, 0, "likes")
	thread := testThread(forum.ID, uuid.New())
	post := testPost(thread.ID, forum.ID, uuid.New(), 1)
	forums.items[forum.ID] = forum
	threads.items[thread.ID] = thread
	posts.items[post.ID] = post
	auth.like[forum.ID] = true

	first, err := service.LikePost(context.Background(), port.LikePostCommand{ActorUserID: actorID, PostID: post.ID})
	if err != nil {
		t.Fatalf("LikePost first error = %v", err)
	}
	second, err := service.LikePost(context.Background(), port.LikePostCommand{ActorUserID: actorID, PostID: post.ID})
	if err != nil {
		t.Fatalf("LikePost second error = %v", err)
	}
	if first.LikeCount != 1 || second.LikeCount != 1 || posts.items[post.ID].LikeCount != 1 || len(interactions.likes) != 1 {
		t.Fatalf("like summaries first=%+v second=%+v count=%d likes=%d, want idempotent count", first, second, posts.items[post.ID].LikeCount, len(interactions.likes))
	}
}

// TestServiceUnlikePostIsIdempotent verifies unlike commands do not drift counts.
func TestServiceUnlikePostIsIdempotent(t *testing.T) {
	service, _, forums, threads, posts, auth, _ := newContentTestService()
	actorID := uuid.New()
	forum := testForum(uuid.New(), nil, 0, "unlikes")
	thread := testThread(forum.ID, uuid.New())
	post := testPost(thread.ID, forum.ID, uuid.New(), 1)
	forums.items[forum.ID] = forum
	threads.items[thread.ID] = thread
	posts.items[post.ID] = post
	auth.like[forum.ID] = true

	if _, err := service.LikePost(context.Background(), port.LikePostCommand{ActorUserID: actorID, PostID: post.ID}); err != nil {
		t.Fatalf("LikePost error = %v", err)
	}
	first, err := service.UnlikePost(context.Background(), port.UnlikePostCommand{ActorUserID: actorID, PostID: post.ID})
	if err != nil {
		t.Fatalf("UnlikePost first error = %v", err)
	}
	second, err := service.UnlikePost(context.Background(), port.UnlikePostCommand{ActorUserID: actorID, PostID: post.ID})
	if err != nil {
		t.Fatalf("UnlikePost second error = %v", err)
	}
	if first.LikeCount != 0 || second.LikeCount != 0 || posts.items[post.ID].LikeCount != 0 {
		t.Fatalf("unlike summaries first=%+v second=%+v count=%d, want idempotent zero", first, second, posts.items[post.ID].LikeCount)
	}
}

// TestServiceLatestPostsUsesVisibleForumsAndCache verifies widget visibility and caching.
func TestServiceLatestPostsUsesVisibleForumsAndCache(t *testing.T) {
	service, _, forums, _, _, auth, interactions := newContentTestService()
	actorID := uuid.New()
	visibleForum := testForum(uuid.New(), nil, 0, "visible")
	hiddenForum := testForum(uuid.New(), nil, 0, "hidden")
	forums.items[visibleForum.ID] = visibleForum
	forums.items[hiddenForum.ID] = hiddenForum
	auth.visible[visibleForum.ID] = true
	interactions.latest = []domain.LatestPostSummary{{ForumID: visibleForum.ID, ThreadID: uuid.New(), PostID: uuid.New(), AuthorUserID: uuid.New(), Sequence: 1, ThreadTitle: "Visible"}}

	first, err := service.ListLatestPosts(context.Background(), actorID, uuid.Nil, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("ListLatestPosts first error = %v", err)
	}
	interactions.latest = nil
	second, err := service.ListLatestPosts(context.Background(), actorID, uuid.Nil, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("ListLatestPosts second error = %v", err)
	}
	if len(first.Items) != 1 || len(second.Items) != 1 || interactions.latestFilters[0].ForumIDs[0] != visibleForum.ID {
		t.Fatalf("latest first=%+v second=%+v filters=%+v, want cached visible latest", first.Items, second.Items, interactions.latestFilters)
	}
}

// TestServiceReadStateRequiresAuthenticatedActor verifies anonymous read state is rejected.
func TestServiceReadStateRequiresAuthenticatedActor(t *testing.T) {
	service, _, _, _, _, _, _ := newContentTestService()

	_, err := service.MarkThreadRead(context.Background(), port.MarkThreadReadCommand{ThreadID: uuid.New(), LastReadPostSequence: 1})
	if !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("MarkThreadRead() error = %v, want %v", err, port.ErrForbidden)
	}
	if err := service.MarkForumRead(context.Background(), port.MarkForumReadCommand{ForumID: uuid.New()}); !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("MarkForumRead() error = %v, want %v", err, port.ErrForbidden)
	}
}

// TestServiceUnreadSummaryUsesVisibleForums verifies unread summaries are visibility-aware.
func TestServiceUnreadSummaryUsesVisibleForums(t *testing.T) {
	service, _, forums, _, _, auth, interactions := newContentTestService()
	actorID := uuid.New()
	forum := testForum(uuid.New(), nil, 0, "unread")
	forums.items[forum.ID] = forum
	auth.visible[forum.ID] = true
	interactions.unread = domain.UnreadSummary{UserID: actorID, UnreadThreadCount: 3, Forums: []domain.ForumUnreadSummary{{ForumID: forum.ID, UnreadThreadCount: 3}}}

	summary, err := service.GetUnreadSummary(context.Background(), actorID)
	if err != nil {
		t.Fatalf("GetUnreadSummary() error = %v", err)
	}
	if summary.UnreadThreadCount != 3 || len(interactions.unreadForumIDs) != 1 || interactions.unreadForumIDs[0] != forum.ID {
		t.Fatalf("summary=%+v visible=%+v, want visible unread summary", summary, interactions.unreadForumIDs)
	}
}

// TestServiceSearchUsesVisibleForums verifies search is visibility-scoped.
func TestServiceSearchUsesVisibleForums(t *testing.T) {
	service, _, forums, _, _, auth, _ := newContentTestService()
	visibleForum := testForum(uuid.New(), nil, 0, "visible-search")
	hiddenForum := testForum(uuid.New(), nil, 0, "hidden-search")
	forums.items[visibleForum.ID] = visibleForum
	forums.items[hiddenForum.ID] = hiddenForum
	auth.visible[visibleForum.ID] = true
	service.operations.(*memoryOperations).search = []domain.SearchResult{{Type: "thread", ForumID: visibleForum.ID, ThreadID: uuid.New(), Title: "Visible"}}

	result, err := service.Search(context.Background(), port.SearchCommand{ActorUserID: uuid.New(), Query: "visible"}, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	operations := service.operations.(*memoryOperations)
	if len(result.Items) != 1 || len(operations.searchFilters) != 1 || len(operations.searchFilters[0].ForumIDs) != 1 || operations.searchFilters[0].ForumIDs[0] != visibleForum.ID {
		t.Fatalf("result=%+v filters=%+v, want only visible forum searched", result, operations.searchFilters)
	}
}

// TestServiceSearchRejectsInvalidQuery verifies search input validation.
func TestServiceSearchRejectsInvalidQuery(t *testing.T) {
	service, _, _, _, _, _, _ := newContentTestService()

	_, err := service.Search(context.Background(), port.SearchCommand{Query: "x"}, pagination.Page{Limit: 10})
	var validation domain.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Search() error = %v, want validation", err)
	}
}

// TestServiceFlushThreadViewsAppliesValidBufferedViews verifies view buffers are sanitized before flush.
func TestServiceFlushThreadViewsAppliesValidBufferedViews(t *testing.T) {
	service, _, _, _, _, _, _ := newContentTestService()
	cache := service.cache.(*memoryCache)
	operations := service.operations.(*memoryOperations)
	validID := uuid.New()
	cache.threadViews[validID.String()] = 3
	cache.threadViews["invalid"] = 5
	cache.threadViews[uuid.NewString()] = -1

	flushed, err := service.FlushThreadViews(context.Background())
	if err != nil {
		t.Fatalf("FlushThreadViews() error = %v", err)
	}
	if flushed != 3 || operations.viewIncrements[validID] != 3 || len(cache.threadViews) != 0 {
		t.Fatalf("flushed=%d increments=%+v cache=%+v, want only valid positive views", flushed, operations.viewIncrements, cache.threadViews)
	}
}

// TestServiceRepairMethodsDelegateToOperations verifies operational use cases remain testable.
func TestServiceRepairMethodsDelegateToOperations(t *testing.T) {
	service, _, _, _, _, _, _ := newContentTestService()
	operations := service.operations.(*memoryOperations)
	operations.report = domain.CounterDriftReport{Mismatches: []domain.CounterDrift{{ObjectType: "forum_thread", ObjectID: uuid.New(), Field: "post_count", Expected: 2, Actual: 1}}}

	stats, err := service.VerifyStats(context.Background())
	if err != nil {
		t.Fatalf("VerifyStats() error = %v", err)
	}
	if len(stats.Mismatches) != 1 || operations.verifyStatsCalls != 1 {
		t.Fatalf("stats=%+v calls=%d, want delegated report", stats, operations.verifyStatsCalls)
	}
	if _, err := service.RebuildStats(context.Background()); err != nil {
		t.Fatalf("RebuildStats() error = %v", err)
	}
	if _, err := service.VerifyLikes(context.Background()); err != nil {
		t.Fatalf("VerifyLikes() error = %v", err)
	}
	if _, err := service.RebuildLikes(context.Background()); err != nil {
		t.Fatalf("RebuildLikes() error = %v", err)
	}
	if operations.rebuildStatsCalls != 1 || operations.verifyLikesCalls != 1 || operations.rebuildLikesCalls != 1 {
		t.Fatalf("operations calls = %+v, want all repair hooks delegated", operations)
	}
}

// newContentTestService creates a forum service with exposed content stores.
func newContentTestService() (Service, *memoryCategories, *memoryForums, *memoryThreads, *memoryPosts, *memoryAuthorizer, *memoryInteractions) {
	categories := &memoryCategories{items: map[uuid.UUID]domain.ForumCategory{}}
	forums := &memoryForums{items: map[uuid.UUID]domain.Forum{}, stats: map[uuid.UUID]domain.ForumStats{}}
	threads := &memoryThreads{items: map[uuid.UUID]domain.Thread{}}
	posts := &memoryPosts{items: map[uuid.UUID]domain.Post{}, revisions: map[uuid.UUID][]domain.PostRevision{}}
	interactions := &memoryInteractions{posts: posts, threads: threads, likes: map[string]domain.PostLike{}, readStates: map[string]domain.ThreadReadState{}}
	operations := &memoryOperations{viewIncrements: map[uuid.UUID]int64{}}
	auth := &memoryAuthorizer{visible: map[uuid.UUID]bool{}, manage: map[uuid.UUID]bool{}, create: map[uuid.UUID]bool{}, reply: map[uuid.UUID]bool{}, like: map[uuid.UUID]bool{}, manageThreads: map[uuid.UUID]bool{}, managePosts: map[uuid.UUID]bool{}}
	service := NewService(Dependencies{Categories: categories, Forums: forums, Threads: threads, Posts: posts, Interactions: interactions, Operations: operations, Assets: &memoryAssets{existing: map[uuid.UUID]bool{}}, Authorizer: auth, Cache: newMemoryCache(), Transactions: noopTx{}})
	return service, categories, forums, threads, posts, auth, interactions
}

// newTestService creates a forum service with in-memory fakes.
func newTestService() (Service, *memoryCategories, *memoryForums, *memoryAuthorizer, *memoryCache) {
	categories := &memoryCategories{items: map[uuid.UUID]domain.ForumCategory{}}
	forums := &memoryForums{items: map[uuid.UUID]domain.Forum{}, stats: map[uuid.UUID]domain.ForumStats{}}
	threads := &memoryThreads{items: map[uuid.UUID]domain.Thread{}}
	posts := &memoryPosts{items: map[uuid.UUID]domain.Post{}, revisions: map[uuid.UUID][]domain.PostRevision{}}
	interactions := &memoryInteractions{posts: posts, threads: threads, likes: map[string]domain.PostLike{}, readStates: map[string]domain.ThreadReadState{}}
	operations := &memoryOperations{viewIncrements: map[uuid.UUID]int64{}}
	auth := &memoryAuthorizer{visible: map[uuid.UUID]bool{}, manage: map[uuid.UUID]bool{}, create: map[uuid.UUID]bool{}, reply: map[uuid.UUID]bool{}, like: map[uuid.UUID]bool{}, manageThreads: map[uuid.UUID]bool{}, managePosts: map[uuid.UUID]bool{}}
	cache := newMemoryCache()
	service := NewService(Dependencies{Categories: categories, Forums: forums, Threads: threads, Posts: posts, Interactions: interactions, Operations: operations, Assets: &memoryAssets{existing: map[uuid.UUID]bool{}}, Authorizer: auth, Cache: cache, Transactions: noopTx{}})
	return service, categories, forums, auth, cache
}

// memoryCategories stores categories in memory.
type memoryCategories struct {
	items map[uuid.UUID]domain.ForumCategory
}

// Create stores a category.
func (repository *memoryCategories) Create(_ context.Context, category domain.ForumCategory) (domain.ForumCategory, error) {
	repository.items[category.ID] = category
	return category, nil
}

// Update stores mutable category fields.
func (repository *memoryCategories) Update(_ context.Context, category domain.ForumCategory, expectedVersion uint64) (domain.ForumCategory, error) {
	current, ok := repository.items[category.ID]
	if !ok {
		return domain.ForumCategory{}, port.ErrNotFound
	}
	if current.Version != expectedVersion {
		return domain.ForumCategory{}, port.ErrPreconditionFailed
	}
	category.Version = expectedVersion + 1
	repository.items[category.ID] = category
	return category, nil
}

// FindByID returns one category.
func (repository *memoryCategories) FindByID(_ context.Context, id uuid.UUID) (domain.ForumCategory, error) {
	item, ok := repository.items[id]
	if !ok {
		return domain.ForumCategory{}, port.ErrNotFound
	}
	return item, nil
}

// List returns matching categories.
func (repository *memoryCategories) List(_ context.Context, filter port.CategoryFilter, _ pagination.Page) (pagination.Result[domain.ForumCategory], error) {
	items := []domain.ForumCategory{}
	for _, item := range repository.items {
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		items = append(items, item)
	}
	return pagination.Result[domain.ForumCategory]{Items: items}, nil
}

// Delete removes a category.
func (repository *memoryCategories) Delete(_ context.Context, id uuid.UUID, _ uint64) error {
	delete(repository.items, id)
	return nil
}

// Reorder updates category order.
func (repository *memoryCategories) Reorder(_ context.Context, items []port.ReorderItem) error {
	for _, item := range items {
		category := repository.items[item.ID]
		category.DisplayOrder = item.DisplayOrder
		repository.items[item.ID] = category
	}
	return nil
}

// memoryForums stores forums in memory.
type memoryForums struct {
	items map[uuid.UUID]domain.Forum
	stats map[uuid.UUID]domain.ForumStats
}

// Create stores a forum.
func (repository *memoryForums) Create(_ context.Context, forum domain.Forum) (domain.Forum, error) {
	repository.items[forum.ID] = forum
	repository.stats[forum.ID] = domain.ForumStats{ForumID: forum.ID}
	return forum, nil
}

// Update stores mutable forum fields.
func (repository *memoryForums) Update(_ context.Context, forum domain.Forum, expectedVersion uint64) (domain.Forum, error) {
	forum.Version = expectedVersion + 1
	repository.items[forum.ID] = forum
	return forum, nil
}

// FindByID returns one forum.
func (repository *memoryForums) FindByID(_ context.Context, id uuid.UUID) (domain.Forum, error) {
	item, ok := repository.items[id]
	if !ok {
		return domain.Forum{}, port.ErrNotFound
	}
	return item, nil
}

// List returns matching forums.
func (repository *memoryForums) List(_ context.Context, _ port.ForumFilter, _ pagination.Page) (pagination.Result[domain.Forum], error) {
	items := make([]domain.Forum, 0, len(repository.items))
	for _, item := range repository.items {
		items = append(items, item)
	}
	return pagination.Result[domain.Forum]{Items: items}, nil
}

// ListTreeForums returns tree forums.
func (repository *memoryForums) ListTreeForums(context.Context) ([]domain.Forum, error) {
	items := make([]domain.Forum, 0, len(repository.items))
	for _, item := range repository.items {
		items = append(items, item)
	}
	return items, nil
}

// ListStats returns stats for ids.
func (repository *memoryForums) ListStats(_ context.Context, ids []uuid.UUID) (map[uuid.UUID]domain.ForumStats, error) {
	result := map[uuid.UUID]domain.ForumStats{}
	for _, id := range ids {
		result[id] = repository.stats[id]
	}
	return result, nil
}

// Move stores moved forum.
func (repository *memoryForums) Move(_ context.Context, forum domain.Forum, _ string, expectedVersion uint64) (domain.Forum, error) {
	forum.Version = expectedVersion + 1
	repository.items[forum.ID] = forum
	return forum, nil
}

// Delete removes a forum.
func (repository *memoryForums) Delete(_ context.Context, id uuid.UUID, _ uint64) error {
	delete(repository.items, id)
	return nil
}

// Reorder updates forum order.
func (repository *memoryForums) Reorder(_ context.Context, items []port.ReorderItem) error {
	for _, item := range items {
		forum := repository.items[item.ID]
		forum.DisplayOrder = item.DisplayOrder
		repository.items[item.ID] = forum
	}
	return nil
}

// memoryAuthorizer stores permission decisions in memory.
type memoryAuthorizer struct {
	visible       map[uuid.UUID]bool
	manage        map[uuid.UUID]bool
	create        map[uuid.UUID]bool
	reply         map[uuid.UUID]bool
	like          map[uuid.UUID]bool
	manageThreads map[uuid.UUID]bool
	managePosts   map[uuid.UUID]bool
}

// VisibleForums returns visible forum IDs.
func (authorizer *memoryAuthorizer) VisibleForums(_ context.Context, _ uuid.UUID, forumIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	result := map[uuid.UUID]bool{}
	for _, id := range forumIDs {
		result[id] = authorizer.visible[id]
	}
	return result, nil
}

// CanManageForum returns management decision.
func (authorizer *memoryAuthorizer) CanManageForum(_ context.Context, _ uuid.UUID, forumID uuid.UUID) (bool, error) {
	return authorizer.manage[forumID], nil
}

// CanCreateThread returns thread creation decision.
func (authorizer *memoryAuthorizer) CanCreateThread(_ context.Context, _ uuid.UUID, forumID uuid.UUID) (bool, error) {
	return authorizer.create[forumID], nil
}

// CanReply returns reply decision.
func (authorizer *memoryAuthorizer) CanReply(_ context.Context, _ uuid.UUID, forumID uuid.UUID) (bool, error) {
	return authorizer.reply[forumID], nil
}

// CanLikePosts returns like decision.
func (authorizer *memoryAuthorizer) CanLikePosts(_ context.Context, _ uuid.UUID, forumID uuid.UUID) (bool, error) {
	return authorizer.like[forumID], nil
}

// CanManageThreads returns thread management decision.
func (authorizer *memoryAuthorizer) CanManageThreads(_ context.Context, _ uuid.UUID, forumID uuid.UUID) (bool, error) {
	return authorizer.manageThreads[forumID], nil
}

// CanManagePosts returns post management decision.
func (authorizer *memoryAuthorizer) CanManagePosts(_ context.Context, _ uuid.UUID, forumID uuid.UUID) (bool, error) {
	return authorizer.managePosts[forumID], nil
}

// memoryThreads stores threads in memory.
type memoryThreads struct {
	items map[uuid.UUID]domain.Thread
}

// Create stores a thread.
func (repository *memoryThreads) Create(_ context.Context, thread domain.Thread) (domain.Thread, error) {
	repository.items[thread.ID] = thread
	return thread, nil
}

// FindByID returns one thread.
func (repository *memoryThreads) FindByID(_ context.Context, id uuid.UUID) (domain.Thread, error) {
	item, ok := repository.items[id]
	if !ok {
		return domain.Thread{}, port.ErrNotFound
	}
	return item, nil
}

// List returns matching threads.
func (repository *memoryThreads) List(_ context.Context, filter port.ThreadFilter, _ pagination.Page) (pagination.Result[domain.Thread], error) {
	items := []domain.Thread{}
	for _, item := range repository.items {
		if item.ForumID == filter.ForumID {
			items = append(items, item)
		}
	}
	return pagination.Result[domain.Thread]{Items: items}, nil
}

// UpdateTitle stores thread title fields.
func (repository *memoryThreads) UpdateTitle(_ context.Context, thread domain.Thread, expectedVersion uint64) (domain.Thread, error) {
	thread.Version = expectedVersion + 1
	repository.items[thread.ID] = thread
	return thread, nil
}

// Delete removes a thread.
func (repository *memoryThreads) Delete(_ context.Context, id uuid.UUID, _ uint64) error {
	delete(repository.items, id)
	return nil
}

// memoryPosts stores posts in memory.
type memoryPosts struct {
	items     map[uuid.UUID]domain.Post
	revisions map[uuid.UUID][]domain.PostRevision
}

// Create stores a post.
func (repository *memoryPosts) Create(_ context.Context, post domain.Post, _ []domain.PostReference) (domain.Post, error) {
	repository.items[post.ID] = post
	return post, nil
}

// FindByID returns one post.
func (repository *memoryPosts) FindByID(_ context.Context, id uuid.UUID) (domain.Post, error) {
	item, ok := repository.items[id]
	if !ok {
		return domain.Post{}, port.ErrNotFound
	}
	return item, nil
}

// List returns matching posts.
func (repository *memoryPosts) List(_ context.Context, filter port.PostFilter, _ pagination.Page) (pagination.Result[domain.Post], error) {
	items := []domain.Post{}
	for _, item := range repository.items {
		if item.ThreadID == filter.ThreadID {
			items = append(items, item)
		}
	}
	return pagination.Result[domain.Post]{Items: items}, nil
}

// NextSequence returns next sequence.
func (repository *memoryPosts) NextSequence(_ context.Context, threadID uuid.UUID) (int64, error) {
	var max int64
	for _, item := range repository.items {
		if item.ThreadID == threadID && item.Sequence > max {
			max = item.Sequence
		}
	}
	return max + 1, nil
}

// UpdateWithRevision stores an updated post and revision.
func (repository *memoryPosts) UpdateWithRevision(_ context.Context, post domain.Post, revision domain.PostRevision, expectedVersion uint64) (domain.Post, error) {
	post.Version = expectedVersion + 1
	repository.items[post.ID] = post
	repository.revisions[post.ID] = append(repository.revisions[post.ID], revision)
	return post, nil
}

// Delete removes a post.
func (repository *memoryPosts) Delete(_ context.Context, id uuid.UUID, _ uint64) error {
	delete(repository.items, id)
	return nil
}

// ListRevisions returns revisions.
func (repository *memoryPosts) ListRevisions(_ context.Context, postID uuid.UUID, _ pagination.Page) (pagination.Result[domain.PostRevision], error) {
	return pagination.Result[domain.PostRevision]{Items: repository.revisions[postID]}, nil
}

// ListReferences returns references.
func (repository *memoryPosts) ListReferences(context.Context, []uuid.UUID) (map[uuid.UUID][]domain.PostReference, error) {
	return map[uuid.UUID][]domain.PostReference{}, nil
}

// memoryInteractions stores interactions in memory.
type memoryInteractions struct {
	posts          *memoryPosts
	threads        *memoryThreads
	likes          map[string]domain.PostLike
	readStates     map[string]domain.ThreadReadState
	latest         []domain.LatestPostSummary
	mostLiked      []domain.MostLikedPost
	unread         domain.UnreadSummary
	latestFilters  []port.LatestPostFilter
	unreadForumIDs []uuid.UUID
}

// LikePost stores an active like once.
func (repository *memoryInteractions) LikePost(_ context.Context, like domain.PostLike) (bool, error) {
	key := likeKey(like.PostID, like.UserID)
	if _, ok := repository.likes[key]; ok {
		return false, nil
	}
	repository.likes[key] = like
	post := repository.posts.items[like.PostID]
	post.LikeCount++
	repository.posts.items[like.PostID] = post
	thread := repository.threads.items[like.ThreadID]
	thread.LikeCount++
	repository.threads.items[like.ThreadID] = thread
	return true, nil
}

// UnlikePost removes an active like once.
func (repository *memoryInteractions) UnlikePost(_ context.Context, postID uuid.UUID, userID uuid.UUID) (bool, error) {
	key := likeKey(postID, userID)
	like, ok := repository.likes[key]
	if !ok {
		return false, nil
	}
	delete(repository.likes, key)
	post := repository.posts.items[postID]
	if post.LikeCount > 0 {
		post.LikeCount--
	}
	repository.posts.items[postID] = post
	thread := repository.threads.items[like.ThreadID]
	if thread.LikeCount > 0 {
		thread.LikeCount--
	}
	repository.threads.items[like.ThreadID] = thread
	return true, nil
}

// LikedByUser reports whether a like exists.
func (repository *memoryInteractions) LikedByUser(_ context.Context, postID uuid.UUID, userID uuid.UUID) (bool, error) {
	_, ok := repository.likes[likeKey(postID, userID)]
	return ok, nil
}

// ListLatestPosts returns latest posts.
func (repository *memoryInteractions) ListLatestPosts(_ context.Context, filter port.LatestPostFilter, _ pagination.Page) (pagination.Result[domain.LatestPostSummary], error) {
	repository.latestFilters = append(repository.latestFilters, filter)
	return pagination.Result[domain.LatestPostSummary]{Items: repository.latest}, nil
}

// ListMostLikedPosts returns most-liked posts.
func (repository *memoryInteractions) ListMostLikedPosts(_ context.Context, _ port.MostLikedFilter, _ pagination.Page) (pagination.Result[domain.MostLikedPost], error) {
	return pagination.Result[domain.MostLikedPost]{Items: repository.mostLiked}, nil
}

// MarkThreadRead stores one read state.
func (repository *memoryInteractions) MarkThreadRead(_ context.Context, state domain.ThreadReadState) error {
	repository.readStates[state.ThreadID.String()] = state
	return nil
}

// MarkForumRead stores read states for all forum threads.
func (repository *memoryInteractions) MarkForumRead(_ context.Context, userID uuid.UUID, forumID uuid.UUID, readAt time.Time) error {
	for _, thread := range repository.threads.items {
		if thread.ForumID != forumID {
			continue
		}
		repository.readStates[thread.ID.String()] = domain.ThreadReadState{ID: uuid.New(), UserID: userID, ForumID: forumID, ThreadID: thread.ID, LastReadPostSequence: thread.VisiblePostCount, LastReadAt: readAt}
	}
	return nil
}

// UnreadSummary returns unread summary.
func (repository *memoryInteractions) UnreadSummary(_ context.Context, _ uuid.UUID, forumIDs []uuid.UUID) (domain.UnreadSummary, error) {
	repository.unreadForumIDs = forumIDs
	return repository.unread, nil
}

// memoryOperations stores operational calls in memory.
type memoryOperations struct {
	search             []domain.SearchResult
	searchFilters      []port.SearchFilter
	report             domain.CounterDriftReport
	viewIncrements     map[uuid.UUID]int64
	verifyStatsCalls   int
	rebuildStatsCalls  int
	verifyLikesCalls   int
	rebuildLikesCalls  int
	applyViewsCallSeen bool
}

// Search returns configured search rows.
func (repository *memoryOperations) Search(_ context.Context, filter port.SearchFilter, _ pagination.Page) (pagination.Result[domain.SearchResult], error) {
	repository.searchFilters = append(repository.searchFilters, filter)
	return pagination.Result[domain.SearchResult]{Items: repository.search}, nil
}

// VerifyStats reports configured stats drift.
func (repository *memoryOperations) VerifyStats(context.Context) (domain.CounterDriftReport, error) {
	repository.verifyStatsCalls++
	return repository.report, nil
}

// RebuildStats reports configured stats drift as repaired.
func (repository *memoryOperations) RebuildStats(context.Context) (domain.CounterDriftReport, error) {
	repository.rebuildStatsCalls++
	report := repository.report
	report.Repaired = true
	return report, nil
}

// VerifyLikes reports configured like drift.
func (repository *memoryOperations) VerifyLikes(context.Context) (domain.CounterDriftReport, error) {
	repository.verifyLikesCalls++
	return repository.report, nil
}

// RebuildLikes reports configured like drift as repaired.
func (repository *memoryOperations) RebuildLikes(context.Context) (domain.CounterDriftReport, error) {
	repository.rebuildLikesCalls++
	report := repository.report
	report.Repaired = true
	return report, nil
}

// ApplyThreadViews stores flushed view increments.
func (repository *memoryOperations) ApplyThreadViews(_ context.Context, increments map[uuid.UUID]int64) error {
	repository.applyViewsCallSeen = true
	for threadID, increment := range increments {
		repository.viewIncrements[threadID] += increment
	}
	return nil
}

// likeKey returns an in-memory like key.
func likeKey(postID uuid.UUID, userID uuid.UUID) string {
	return postID.String() + ":" + userID.String()
}

// memoryAssets resolves attachment IDs in memory.
type memoryAssets struct {
	existing map[uuid.UUID]bool
}

// AssetExists reports whether an asset exists.
func (resolver *memoryAssets) AssetExists(_ context.Context, id uuid.UUID) (bool, error) {
	return resolver.existing[id], nil
}

// memoryCache stores trees in memory.
type memoryCache struct {
	items       map[string]domain.ForumTree
	latest      map[string]pagination.Result[domain.LatestPostSummary]
	mostLiked   map[string]pagination.Result[domain.MostLikedPost]
	threadViews map[string]int64
	sets        int
}

// newMemoryCache creates a memory read cache.
func newMemoryCache() *memoryCache {
	return &memoryCache{items: map[string]domain.ForumTree{}, latest: map[string]pagination.Result[domain.LatestPostSummary]{}, mostLiked: map[string]pagination.Result[domain.MostLikedPost]{}, threadViews: map[string]int64{}}
}

// GetTree returns a cached tree.
func (cache *memoryCache) GetTree(_ context.Context, key string) (domain.ForumTree, bool, error) {
	tree, ok := cache.items[key]
	return tree, ok, nil
}

// SetTree stores a tree.
func (cache *memoryCache) SetTree(_ context.Context, key string, tree domain.ForumTree, _ time.Duration) error {
	cache.items[key] = tree
	cache.sets++
	return nil
}

// ClearTree clears trees.
func (cache *memoryCache) ClearTree(context.Context) error {
	cache.items = map[string]domain.ForumTree{}
	return nil
}

// GetLatestPosts returns cached latest posts.
func (cache *memoryCache) GetLatestPosts(_ context.Context, key string) (pagination.Result[domain.LatestPostSummary], bool, error) {
	result, ok := cache.latest[key]
	return result, ok, nil
}

// SetLatestPosts stores latest posts.
func (cache *memoryCache) SetLatestPosts(_ context.Context, key string, result pagination.Result[domain.LatestPostSummary], _ time.Duration) error {
	cache.latest[key] = result
	return nil
}

// ClearLatestPosts clears latest posts.
func (cache *memoryCache) ClearLatestPosts(context.Context) error {
	cache.latest = map[string]pagination.Result[domain.LatestPostSummary]{}
	return nil
}

// GetMostLikedPosts returns cached most-liked posts.
func (cache *memoryCache) GetMostLikedPosts(_ context.Context, key string) (pagination.Result[domain.MostLikedPost], bool, error) {
	result, ok := cache.mostLiked[key]
	return result, ok, nil
}

// SetMostLikedPosts stores most-liked posts.
func (cache *memoryCache) SetMostLikedPosts(_ context.Context, key string, result pagination.Result[domain.MostLikedPost], _ time.Duration) error {
	cache.mostLiked[key] = result
	return nil
}

// ClearMostLikedPosts clears most-liked posts.
func (cache *memoryCache) ClearMostLikedPosts(context.Context) error {
	cache.mostLiked = map[string]pagination.Result[domain.MostLikedPost]{}
	return nil
}

// IncrementThreadView buffers a thread view.
func (cache *memoryCache) IncrementThreadView(_ context.Context, threadID string) error {
	cache.threadViews[threadID]++
	return nil
}

// DrainThreadViews returns and clears buffered thread views.
func (cache *memoryCache) DrainThreadViews(context.Context) (map[string]int64, error) {
	views := cache.threadViews
	cache.threadViews = map[string]int64{}
	return views, nil
}

// ClearAll clears read caches.
func (cache *memoryCache) ClearAll(context.Context) error {
	cache.items = map[string]domain.ForumTree{}
	cache.latest = map[string]pagination.Result[domain.LatestPostSummary]{}
	cache.mostLiked = map[string]pagination.Result[domain.MostLikedPost]{}
	return nil
}

// noopTx runs work without a real transaction.
type noopTx struct{}

// WithinTx runs fn.
func (noopTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// testCategory returns a category.
func testCategory() domain.ForumCategory {
	return domain.ForumCategory{ID: uuid.New(), Key: "official", Name: "Official", Status: domain.CategoryStatusActive, Version: 1}
}

// testForum returns a forum.
func testForum(categoryID uuid.UUID, parentID *uuid.UUID, order int, key string) domain.Forum {
	id := uuid.New()
	path := "/" + id.String() + "/"
	return domain.Forum{ID: id, CategoryID: categoryID, ParentForumID: parentID, Kind: domain.ForumKindDiscussion, Key: domain.Key(key), Slug: domain.Slug(key), Name: key, DisplayOrder: order, Path: path, ThreadVisibilityMode: domain.ThreadVisibilityAllThreads, DefaultThreadStatus: domain.ThreadStatusOpen, Status: domain.ForumStatusActive, Version: 1}
}

// testThread returns a thread.
func testThread(forumID uuid.UUID, authorID uuid.UUID) domain.Thread {
	postID := uuid.New()
	return domain.Thread{ID: uuid.New(), ForumID: forumID, AuthorUserID: authorID, OpenerPostID: postID, LatestPostID: postID, LatestPostAuthorUserID: authorID, LatestPostAt: time.Now().UTC(), Title: "A thread", Slug: "a-thread", Status: domain.ThreadStatusOpen, StickyState: domain.StickyStateNormal, PostCount: 1, VisiblePostCount: 1, Version: 1}
}

// testPost returns a post.
func testPost(threadID uuid.UUID, forumID uuid.UUID, authorID uuid.UUID, sequence int64) domain.Post {
	now := time.Now().UTC()
	return domain.Post{ID: uuid.New(), ThreadID: threadID, ForumID: forumID, AuthorUserID: authorID, Sequence: sequence, Status: domain.PostStatusVisible, ContentFormat: domain.ContentFormatProseMirror, ContentDocumentJSON: []byte(`{"type":"doc"}`), ContentText: "Original", Version: 1, CreatedAt: now, UpdatedAt: now}
}
