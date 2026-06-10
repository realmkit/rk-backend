package content

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/adapter/http/shared"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// TestContentRoutesExerciseHandlers verifies content route wiring and response codes.
func TestContentRoutesExerciseHandlers(t *testing.T) {
	app := fiber.New()
	Register(app.Group("/api/v1"), Services{
		Content:     contentService{},
		Interaction: interactionService{},
		Operations:  operationsService{},
	})
	userID := uuid.NewString()
	for _, route := range contentRoutes() {
		req, _ := http.NewRequest(route.method, route.path, bytes.NewBufferString(route.body))
		req.Header.Set(headers.ContentType, "application/json")
		req.Header.Set(headers.Accept, "application/json")
		req.Header.Set(shared.CurrentUserIDHeader, userID)
		if route.idempotent {
			req.Header.Set(headers.IdempotencyKey, route.path)
		}
		if route.versioned {
			req.Header.Set(headers.IfMatch, `"1"`)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s Test() error = %v", route.method, route.path, err)
		}
		if resp.StatusCode != route.status {
			t.Fatalf("%s %s status = %d, want %d", route.method, route.path, resp.StatusCode, route.status)
		}
	}
}

// contentRoute describes one route test case.
type contentRoute struct {
	method     string
	path       string
	body       string
	status     int
	idempotent bool
	versioned  bool
}

// contentRoutes returns route cases for content handlers.
func contentRoutes() []contentRoute {
	forumID := uuid.NewString()
	threadID := uuid.NewString()
	postID := uuid.NewString()
	return []contentRoute{
		{method: http.MethodGet, path: "/api/v1/forums/latest-posts", status: fiber.StatusOK},
		{method: http.MethodGet, path: "/api/v1/forums/search?query=hello", status: fiber.StatusOK},
		{method: http.MethodGet, path: "/api/v1/forums/unread-summary", status: fiber.StatusOK},
		{method: http.MethodGet, path: "/api/v1/forums/" + forumID + "/threads", status: fiber.StatusOK},
		{method: http.MethodPost, path: "/api/v1/forums/" + forumID + "/threads", body: `{"title":"Hello","content_document_json":{"type":"doc"}}`, status: fiber.StatusCreated, idempotent: true},
		{method: http.MethodGet, path: "/api/v1/forums/" + forumID + "/latest-posts", status: fiber.StatusOK},
		{method: http.MethodGet, path: "/api/v1/forums/" + forumID + "/posts/most-liked", status: fiber.StatusOK},
		{method: http.MethodGet, path: "/api/v1/forums/" + forumID + "/search?query=hello", status: fiber.StatusOK},
		{method: http.MethodPost, path: "/api/v1/forums/" + forumID + "/read", status: fiber.StatusNoContent, idempotent: true},
		{method: http.MethodGet, path: "/api/v1/threads/" + threadID, status: fiber.StatusOK},
		{method: http.MethodPatch, path: "/api/v1/threads/" + threadID, body: `{"title":"Next"}`, status: fiber.StatusOK, idempotent: true, versioned: true},
		{method: http.MethodDelete, path: "/api/v1/threads/" + threadID, status: fiber.StatusNoContent, idempotent: true, versioned: true},
		{method: http.MethodGet, path: "/api/v1/threads/" + threadID + "/posts", status: fiber.StatusOK},
		{method: http.MethodPost, path: "/api/v1/threads/" + threadID + "/posts", body: `{"content_document_json":{"type":"doc"}}`, status: fiber.StatusCreated, idempotent: true},
		{method: http.MethodPost, path: "/api/v1/threads/" + threadID + "/read", body: `{}`, status: fiber.StatusOK, idempotent: true},
		{method: http.MethodGet, path: "/api/v1/posts/" + postID, status: fiber.StatusOK},
		{method: http.MethodPatch, path: "/api/v1/posts/" + postID, body: `{"content_document_json":{"type":"doc"}}`, status: fiber.StatusOK, idempotent: true, versioned: true},
		{method: http.MethodDelete, path: "/api/v1/posts/" + postID, status: fiber.StatusNoContent, idempotent: true, versioned: true},
		{method: http.MethodPut, path: "/api/v1/posts/" + postID + "/like", status: fiber.StatusOK, idempotent: true},
		{method: http.MethodDelete, path: "/api/v1/posts/" + postID + "/like", status: fiber.StatusOK, idempotent: true},
		{method: http.MethodGet, path: "/api/v1/posts/" + postID + "/revisions", status: fiber.StatusOK},
	}
}

// contentService is a test content service.
type contentService struct{}

// CreateThread creates a test thread.
func (contentService) CreateThread(context.Context, port.CreateThreadCommand) (domain.Thread, domain.Post, error) {
	thread := domain.Thread{ID: uuid.New(), Version: 1}
	post := domain.Post{ID: uuid.New(), Version: 1}
	return thread, post, nil
}

// GetThread returns a test thread.
func (contentService) GetThread(context.Context, uuid.UUID, uuid.UUID) (domain.Thread, error) {
	return domain.Thread{ID: uuid.New(), Version: 1}, nil
}

// ListThreads returns a test thread page.
func (contentService) ListThreads(context.Context, uuid.UUID, port.ThreadFilter, pagination.Page) (pagination.Result[domain.Thread], error) {
	return pagination.Result[domain.Thread]{Items: []domain.Thread{}}, nil
}

// UpdateThreadTitle returns an updated test thread.
func (contentService) UpdateThreadTitle(context.Context, port.UpdateThreadTitleCommand) (domain.Thread, error) {
	return domain.Thread{ID: uuid.New(), Version: 2}, nil
}

// DeleteThread deletes a test thread.
func (contentService) DeleteThread(context.Context, port.DeleteThreadCommand) error { return nil }

// CreateReply creates a test reply.
func (contentService) CreateReply(context.Context, port.CreateReplyCommand) (domain.Post, error) {
	return domain.Post{ID: uuid.New(), Version: 1}, nil
}

// ListPosts returns a test post page.
func (contentService) ListPosts(context.Context, uuid.UUID, port.PostFilter, pagination.Page) (pagination.Result[domain.Post], error) {
	return pagination.Result[domain.Post]{Items: []domain.Post{}}, nil
}

// GetPost returns a test post.
func (contentService) GetPost(context.Context, uuid.UUID, uuid.UUID) (domain.Post, error) {
	return domain.Post{ID: uuid.New(), Version: 1}, nil
}

// UpdatePost returns an updated test post.
func (contentService) UpdatePost(context.Context, port.UpdatePostCommand) (domain.Post, error) {
	return domain.Post{ID: uuid.New(), Version: 2}, nil
}

// DeletePost deletes a test post.
func (contentService) DeletePost(context.Context, port.DeletePostCommand) error { return nil }

// ListPostRevisions returns a test revision page.
func (contentService) ListPostRevisions(context.Context, uuid.UUID, uuid.UUID, pagination.Page) (pagination.Result[domain.PostRevision], error) {
	return pagination.Result[domain.PostRevision]{Items: []domain.PostRevision{}}, nil
}

// interactionService is a test interaction service.
type interactionService struct{}

// LikePost returns a like summary.
func (interactionService) LikePost(context.Context, port.LikePostCommand) (domain.PostLikeSummary, error) {
	return domain.PostLikeSummary{}, nil
}

// UnlikePost returns a like summary.
func (interactionService) UnlikePost(context.Context, port.UnlikePostCommand) (domain.PostLikeSummary, error) {
	return domain.PostLikeSummary{}, nil
}

// ListLatestPosts returns latest posts.
func (interactionService) ListLatestPosts(context.Context, uuid.UUID, uuid.UUID, pagination.Page) (pagination.Result[domain.LatestPostSummary], error) {
	return pagination.Result[domain.LatestPostSummary]{Items: []domain.LatestPostSummary{}}, nil
}

// ListMostLikedPosts returns most-liked posts.
func (interactionService) ListMostLikedPosts(context.Context, uuid.UUID, uuid.UUID, pagination.Page) (pagination.Result[domain.MostLikedPost], error) {
	return pagination.Result[domain.MostLikedPost]{Items: []domain.MostLikedPost{}}, nil
}

// MarkThreadRead marks a thread read.
func (interactionService) MarkThreadRead(context.Context, port.MarkThreadReadCommand) (domain.ThreadReadState, error) {
	return domain.ThreadReadState{ID: uuid.New()}, nil
}

// MarkForumRead marks a forum read.
func (interactionService) MarkForumRead(context.Context, port.MarkForumReadCommand) error { return nil }

// GetUnreadSummary returns an unread summary.
func (interactionService) GetUnreadSummary(context.Context, uuid.UUID) (domain.UnreadSummary, error) {
	return domain.UnreadSummary{}, nil
}

// operationsService is a test operations service.
type operationsService struct{}

// Search returns search rows.
func (operationsService) Search(context.Context, port.SearchCommand, pagination.Page) (pagination.Result[domain.SearchResult], error) {
	return pagination.Result[domain.SearchResult]{Items: []domain.SearchResult{}}, nil
}

// VerifyStats reports drift.
func (operationsService) VerifyStats(context.Context) (domain.CounterDriftReport, error) {
	return domain.CounterDriftReport{}, nil
}

// RebuildStats repairs drift.
func (operationsService) RebuildStats(context.Context) (domain.CounterDriftReport, error) {
	return domain.CounterDriftReport{}, nil
}

// VerifyLikes reports like drift.
func (operationsService) VerifyLikes(context.Context) (domain.CounterDriftReport, error) {
	return domain.CounterDriftReport{}, nil
}

// RebuildLikes repairs like drift.
func (operationsService) RebuildLikes(context.Context) (domain.CounterDriftReport, error) {
	return domain.CounterDriftReport{}, nil
}

// FlushThreadViews flushes views.
func (operationsService) FlushThreadViews(context.Context) (int64, error) { return 0, nil }

// ClearReadCache clears read caches.
func (operationsService) ClearReadCache(context.Context) error { return nil }
