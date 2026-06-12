package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// TestTreeAllowsAnonymous verifies the tree route accepts anonymous callers.
func TestTreeAllowsAnonymous(t *testing.T) {
	app := newTestApp(httpService{tree: domain.ForumTree{Categories: []domain.CategoryNode{}}})
	req, _ := http.NewRequest(http.MethodGet, "/forums/tree", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}

// TestCreateCategoryRequiresIdempotency verifies create command headers.
func TestCreateCategoryRequiresIdempotency(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodPost, "/forum-categories", bytes.NewBufferString(`{}`))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.Accept, "application/json")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

// TestCreateCategoryReturnsCreatedETag verifies successful category creation response metadata.
func TestCreateCategoryReturnsCreatedETag(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodPost, "/forum-categories", bytes.NewBufferString(`{"key":"official","name":"Official"}`))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.Accept, "application/json")
	req.Header.Set(headers.IdempotencyKey, "create-category")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusCreated)
	}
	if resp.Header.Get(headers.ETag) != `"1"` {
		t.Fatalf("ETag = %q, want %q", resp.Header.Get(headers.ETag), `"1"`)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body["name"] != "Official" {
		t.Fatalf("body = %+v, want created category", body)
	}
}

// TestCreateCategoryMapsForbidden verifies permission errors become problem responses.
func TestCreateCategoryMapsForbidden(t *testing.T) {
	app := newTestApp(httpService{err: port.ErrForbidden})
	req, _ := http.NewRequest(http.MethodPost, "/forum-categories", bytes.NewBufferString(`{"key":"official","name":"Official"}`))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.Accept, "application/json")
	req.Header.Set(headers.IdempotencyKey, "create-category")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusForbidden)
	}
}

// TestTreeRejectsInvalidOptionalUser verifies optional anonymous auth parsing.
func TestTreeRejectsInvalidOptionalUser(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodGet, "/forums/tree", nil)
	req.Header.Set(currentUserIDHeader, "not-a-uuid")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

// TestUpdateForumRequiresIfMatch verifies optimistic concurrency headers.
func TestUpdateForumRequiresIfMatch(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodPatch, "/forums/"+uuid.NewString(), bytes.NewBufferString(`{}`))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.Accept, "application/json")
	req.Header.Set(headers.IdempotencyKey, "key")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusPreconditionRequired {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusPreconditionRequired)
	}
}

// TestCreateThreadReturnsCreated verifies thread creation route response shape.
func TestCreateThreadReturnsCreated(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(
		http.MethodPost,
		"/forums/"+uuid.NewString()+"/threads",
		bytes.NewBufferString(`{"title":"Hello world","slug":"hello-world","content_document_json":{"type":"doc"},"content_text":"Hello"}`),
	)
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.Accept, "application/json")
	req.Header.Set(headers.IdempotencyKey, "create-thread")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusCreated)
	}
	if resp.Header.Get(headers.ETag) != `"1"` {
		t.Fatalf("ETag = %q, want %q", resp.Header.Get(headers.ETag), `"1"`)
	}
}

// TestCreateReplyRequiresIdempotency verifies reply command headers.
func TestCreateReplyRequiresIdempotency(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(
		http.MethodPost,
		"/threads/"+uuid.NewString()+"/posts",
		bytes.NewBufferString(`{"content_document_json":{"type":"doc"},"content_text":"Reply"}`),
	)
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.Accept, "application/json")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

// TestUpdatePostRequiresIfMatch verifies post edit concurrency headers.
func TestUpdatePostRequiresIfMatch(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(
		http.MethodPatch,
		"/posts/"+uuid.NewString(),
		bytes.NewBufferString(`{"content_document_json":{"type":"doc"},"content_text":"Edit"}`),
	)
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.Accept, "application/json")
	req.Header.Set(headers.IdempotencyKey, "edit-post")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusPreconditionRequired {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusPreconditionRequired)
	}
}

// TestListThreadsReturnsOK verifies thread list route.
func TestListThreadsReturnsOK(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodGet, "/forums/"+uuid.NewString()+"/threads", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}

// TestListPostsReturnsOK verifies post page route.
func TestListPostsReturnsOK(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodGet, "/threads/"+uuid.NewString()+"/posts", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}

// TestGetPostSetsETag verifies direct post response metadata.
func TestGetPostSetsETag(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodGet, "/posts/"+uuid.NewString(), nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
	if resp.Header.Get(headers.ETag) != `"1"` {
		t.Fatalf("ETag = %q, want %q", resp.Header.Get(headers.ETag), `"1"`)
	}
}

// TestListPostRevisionsRequiresUser verifies revision route requires authentication.
func TestListPostRevisionsRequiresUser(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodGet, "/posts/"+uuid.NewString()+"/revisions", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

// TestLikePostRequiresIdempotency verifies like command headers.
func TestLikePostRequiresIdempotency(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodPut, "/posts/"+uuid.NewString()+"/like", nil)
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

// TestLikePostReturnsSummary verifies successful like route.
func TestLikePostReturnsSummary(t *testing.T) {
	app := newTestApp(httpService{})
	postID := uuid.New()
	req, _ := http.NewRequest(http.MethodPut, "/posts/"+postID.String()+"/like", nil)
	req.Header.Set(headers.IdempotencyKey, "like")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
	var summary domain.PostLikeSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if !summary.LikedByActor || summary.LikeCount != 1 {
		t.Fatalf("summary = %+v, want liked summary", summary)
	}
}

// TestUnlikePostReturnsSummary verifies successful unlike route.
func TestUnlikePostReturnsSummary(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodDelete, "/posts/"+uuid.NewString()+"/like", nil)
	req.Header.Set(headers.IdempotencyKey, "unlike")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}

// TestLatestPostsRoutesReturnOK verifies latest widget routes.
func TestLatestPostsRoutesReturnOK(t *testing.T) {
	app := newTestApp(httpService{})
	for _, path := range []string{"/forums/latest-posts", "/forums/" + uuid.NewString() + "/latest-posts"} {
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test(%s) error = %v", path, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status for %s = %d, want %d", path, resp.StatusCode, fiber.StatusOK)
		}
	}
}

// TestMostLikedPostsRouteReturnsOK verifies most-liked widget route.
func TestMostLikedPostsRouteReturnsOK(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodGet, "/forums/"+uuid.NewString()+"/posts/most-liked", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}

// TestSearchRoutesReturnOK verifies global and forum-scoped search routes.
func TestSearchRoutesReturnOK(t *testing.T) {
	app := newTestApp(httpService{})
	for _, path := range []string{"/forums/search?query=thread", "/forums/" + uuid.NewString() + "/search?q=thread"} {
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test(%s) error = %v", path, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status for %s = %d, want %d", path, resp.StatusCode, fiber.StatusOK)
		}
	}
}

// TestMarkThreadReadRequiresUser verifies read state auth.
func TestMarkThreadReadRequiresUser(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodPost, "/threads/"+uuid.NewString()+"/read", nil)
	req.Header.Set(headers.IdempotencyKey, "read")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

// TestMarkForumReadReturnsNoContent verifies forum read route.
func TestMarkForumReadReturnsNoContent(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodPost, "/forums/"+uuid.NewString()+"/read", nil)
	req.Header.Set(headers.IdempotencyKey, "read-forum")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}
}

// TestUnreadSummaryReturnsOK verifies unread summary route.
func TestUnreadSummaryReturnsOK(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodGet, "/forums/unread-summary", nil)
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}

// TestForumHTTPWriteAndReadLifecycleRoutesExerciseHandlers verifies route handler coverage for structure and content.
func TestForumHTTPWriteAndReadLifecycleRoutesExerciseHandlers(t *testing.T) {
	app := newTestApp(httpService{})
	id := uuid.NewString()
	categoryBody := `{"key":"official","name":"Official","status":"active"}`
	forumBody := `{"category_id":"` + uuid.NewString() + `","kind":"discussion","key":"news","slug":"news","name":"News","thread_visibility_mode":"all_threads","default_thread_status":"open","status":"active"}`
	reorderBody := `{"items":[{"id":"` + uuid.NewString() + `","display_order":1}]}`
	moveBody := `{"category_id":"` + uuid.NewString() + `","display_order":2}`
	postBody := `{"content_document_json":{"type":"doc"},"content_text":"Body"}`
	tests := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{method: http.MethodGet, path: "/forum-categories/" + id, status: fiber.StatusOK},
		{method: http.MethodGet, path: "/forum-categories", status: fiber.StatusOK},
		{method: http.MethodPatch, path: "/forum-categories/" + id, body: categoryBody, status: fiber.StatusOK},
		{method: http.MethodDelete, path: "/forum-categories/" + id, status: fiber.StatusNoContent},
		{method: http.MethodPost, path: "/forum-categories/reorder", body: reorderBody, status: fiber.StatusNoContent},
		{method: http.MethodPost, path: "/forums", body: forumBody, status: fiber.StatusCreated},
		{method: http.MethodGet, path: "/forums/" + id, status: fiber.StatusOK},
		{
			method: http.MethodGet,
			path:   "/forums?category_id=" + uuid.NewString() + "&parent_forum_id=" + uuid.NewString() + "&status=active",
			status: fiber.StatusOK,
		},
		{method: http.MethodPatch, path: "/forums/" + id, body: forumBody, status: fiber.StatusOK},
		{method: http.MethodGet, path: "/forums/" + id + "/settings", status: fiber.StatusOK},
		{
			method: http.MethodPatch,
			path:   "/forums/" + id + "/settings",
			body:   `{"kind":"discussion","thread_visibility_mode":"own_threads","max_sticky_threads":3,"default_thread_status":"open","author_post_edit_window_seconds":120,"author_post_delete_window_seconds":60}`,
			status: fiber.StatusOK,
		},
		{method: http.MethodGet, path: "/forums/" + id + "/permissions", status: fiber.StatusOK},
		{
			method: http.MethodPut,
			path:   "/forums/" + id + "/permissions",
			body:   `{"viewers":[{"subject_type":"public"}]}`,
			status: fiber.StatusNoContent,
		},
		{
			method: http.MethodPost,
			path:   "/forums/" + id + "/permissions/simulate",
			body:   `{"permission":"forums.view","object_type":"forum"}`,
			status: fiber.StatusOK,
		},
		{method: http.MethodPost, path: "/forums/" + id + "/move", body: moveBody, status: fiber.StatusOK},
		{method: http.MethodDelete, path: "/forums/" + id, status: fiber.StatusNoContent},
		{method: http.MethodPost, path: "/forums/reorder", body: reorderBody, status: fiber.StatusNoContent},
		{method: http.MethodGet, path: "/forums/search?query=thread", status: fiber.StatusOK},
		{method: http.MethodGet, path: "/forums/" + id + "/search?query=thread", status: fiber.StatusOK},
		{method: http.MethodGet, path: "/threads/" + id, status: fiber.StatusOK},
		{method: http.MethodPatch, path: "/threads/" + id, body: `{"title":"Updated","slug":"updated"}`, status: fiber.StatusOK},
		{method: http.MethodDelete, path: "/threads/" + id, status: fiber.StatusNoContent},
		{method: http.MethodPost, path: "/threads/" + id + "/posts", body: postBody, status: fiber.StatusCreated},
		{method: http.MethodDelete, path: "/posts/" + id, status: fiber.StatusNoContent},
		{method: http.MethodGet, path: "/posts/" + id + "/revisions", status: fiber.StatusOK},
		{method: http.MethodPost, path: "/threads/" + id + "/read", body: `{"last_read_post_sequence":1}`, status: fiber.StatusOK},
	}
	for _, tt := range tests {
		req, _ := http.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
		req.Header.Set(headers.Accept, "application/json")
		req.Header.Set(headers.ContentType, "application/json")
		req.Header.Set(headers.IdempotencyKey, "test-key")
		req.Header.Set(headers.IfMatch, `"1"`)
		req.Header.Set(currentUserIDHeader, uuid.NewString())
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test(%s %s) error = %v", tt.method, tt.path, err)
		}
		if resp.StatusCode != tt.status {
			t.Fatalf("%s %s status = %d, want %d", tt.method, tt.path, resp.StatusCode, tt.status)
		}
	}
}

// TestForumHTTPProblemMappingsExerciseSupport verifies application errors become problem responses.
func TestForumHTTPProblemMappingsExerciseSupport(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{err: port.ErrNotFound, status: fiber.StatusNotFound},
		{err: port.ErrPreconditionFailed, status: fiber.StatusPreconditionFailed},
		{err: port.ErrConflict, status: fiber.StatusConflict},
		{err: port.ErrInvalidMove, status: fiber.StatusConflict},
	}
	for _, tt := range tests {
		app := newTestApp(httpService{err: tt.err})
		req, _ := http.NewRequest(http.MethodGet, "/forums/"+uuid.NewString(), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test() error = %v", err)
		}
		if resp.StatusCode != tt.status {
			t.Fatalf("status for %v = %d, want %d", tt.err, resp.StatusCode, tt.status)
		}
	}
}

// newTestApp creates a Fiber app with forum routes.
func newTestApp(service httpService) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(headers.Middleware())
	useTestPrincipal(app)
	v1 := app
	Register(v1, Services{
		Structure:   service,
		Content:     service,
		Interaction: service,
		Operations:  service,
		Admin:       service,
	})
	return app
}

// httpService is a forum service test double.
type httpService struct {
	tree domain.ForumTree
	err  error
}

// CreateCategory creates a category.
func (service httpService) CreateCategory(context.Context, port.CreateCategoryCommand) (domain.ForumCategory, error) {
	return domain.ForumCategory{
		ID:      uuid.New(),
		Key:     "official",
		Name:    "Official",
		Status:  domain.CategoryStatusActive,
		Version: 1,
	}, service.err
}

// UpdateCategory updates a category.
func (service httpService) UpdateCategory(context.Context, port.UpdateCategoryCommand) (domain.ForumCategory, error) {
	return domain.ForumCategory{
		ID:      uuid.New(),
		Key:     "official",
		Name:    "Official",
		Status:  domain.CategoryStatusActive,
		Version: 2,
	}, service.err
}

// GetCategory returns one category.
func (service httpService) GetCategory(context.Context, uuid.UUID) (domain.ForumCategory, error) {
	return domain.ForumCategory{
		ID:      uuid.New(),
		Key:     "official",
		Name:    "Official",
		Status:  domain.CategoryStatusActive,
		Version: 1,
	}, service.err
}

// ListCategories lists categories.
func (service httpService) ListCategories(
	context.Context,
	port.CategoryFilter,
	pagination.Page,
) (pagination.Result[domain.ForumCategory], error) {
	return pagination.Result[domain.ForumCategory]{
		Items: []domain.ForumCategory{{ID: uuid.New(), Key: "official", Name: "Official", Status: domain.CategoryStatusActive, Version: 1}},
	}, service.err
}

// DeleteCategory deletes a category.
func (service httpService) DeleteCategory(context.Context, port.DeleteCategoryCommand) error {
	return service.err
}

// ReorderCategories reorders categories.
func (service httpService) ReorderCategories(context.Context, port.ReorderCategoriesCommand) error {
	return service.err
}

// CreateForum creates a forum.
func (service httpService) CreateForum(context.Context, port.CreateForumCommand) (domain.Forum, error) {
	return domain.Forum{ID: uuid.New(), Key: "news", Name: "News", Version: 1}, service.err
}

// UpdateForum updates a forum.
func (service httpService) UpdateForum(context.Context, port.UpdateForumCommand) (domain.Forum, error) {
	return domain.Forum{ID: uuid.New(), Key: "news", Name: "News", Version: 2}, service.err
}

// GetForumSettings returns forum settings.
func (service httpService) GetForumSettings(context.Context, uuid.UUID, uuid.UUID) (domain.ForumSettings, error) {
	return domain.ForumSettings{
		ForumID:                       uuid.New(),
		Kind:                          domain.ForumKindDiscussion,
		ThreadVisibilityMode:          domain.ThreadVisibilityAllThreads,
		DefaultThreadStatus:           domain.ThreadStatusOpen,
		AuthorPostEditWindowSeconds:   600,
		AuthorPostDeleteWindowSeconds: 300,
		Version:                       1,
	}, service.err
}

// UpdateForumSettings updates forum settings.
func (service httpService) UpdateForumSettings(context.Context, port.UpdateForumSettingsCommand) (domain.ForumSettings, error) {
	return domain.ForumSettings{
		ForumID:                       uuid.New(),
		Kind:                          domain.ForumKindDiscussion,
		ThreadVisibilityMode:          domain.ThreadVisibilityOwnThreads,
		DefaultThreadStatus:           domain.ThreadStatusOpen,
		AuthorPostEditWindowSeconds:   120,
		AuthorPostDeleteWindowSeconds: 60,
		Version:                       2,
	}, service.err
}

// GetForumPermissionSettings returns forum permission settings.
func (service httpService) GetForumPermissionSettings(context.Context, uuid.UUID, uuid.UUID) (domain.ForumPermissionSettings, error) {
	return domain.ForumPermissionSettings{
		ForumID: uuid.New(),
		Viewers: []domain.ForumPermissionGrant{
			{SubjectType: domain.PermissionSubjectPublic, SubjectID: domain.PublicPermissionSubjectID()},
		},
	}, service.err
}

// UpdateForumPermissionSettings updates forum permission settings.
func (service httpService) UpdateForumPermissionSettings(context.Context, port.UpdateForumPermissionSettingsCommand) error {
	return service.err
}

// SimulateForumPermission simulates one forum permission.
func (service httpService) SimulateForumPermission(
	context.Context,
	port.SimulateForumPermissionCommand,
) (domain.ForumPermissionSimulationResult, error) {
	return domain.ForumPermissionSimulationResult{
		Allowed:          true,
		Reason:           "matched_relation",
		Permission:       "forums.view",
		ObjectType:       "forum",
		ObjectID:         uuid.New(),
		MatchedRelation:  "viewer",
		CheckedRelations: []string{"viewer"},
	}, service.err
}

// MoveForum moves a forum.
func (service httpService) MoveForum(context.Context, port.MoveForumCommand) (domain.Forum, error) {
	return domain.Forum{ID: uuid.New(), Key: "news", Name: "News", Version: 2}, service.err
}

// GetForum returns one forum.
func (service httpService) GetForum(context.Context, uuid.UUID) (domain.Forum, error) {
	return domain.Forum{ID: uuid.New(), Key: "news", Name: "News", Version: 1}, service.err
}

// ListForums lists forums.
func (service httpService) ListForums(context.Context, port.ForumFilter, pagination.Page) (pagination.Result[domain.Forum], error) {
	return pagination.Result[domain.Forum]{Items: []domain.Forum{{ID: uuid.New(), Key: "news", Name: "News", Version: 1}}}, service.err
}

// DeleteForum deletes a forum.
func (service httpService) DeleteForum(context.Context, port.DeleteForumCommand) error {
	return service.err
}

// ReorderForums reorders forums.
func (service httpService) ReorderForums(context.Context, port.ReorderForumsCommand) error {
	return service.err
}

// Tree returns the visible forum tree.
func (service httpService) Tree(context.Context, uuid.UUID) (domain.ForumTree, error) {
	return service.tree, service.err
}

// CreateThread creates a thread.
func (service httpService) CreateThread(context.Context, port.CreateThreadCommand) (domain.Thread, domain.Post, error) {
	return domain.Thread{ID: uuid.New(), Title: "Thread", Version: 1}, domain.Post{ID: uuid.New(), Version: 1}, service.err
}

// GetThread returns one thread.
func (service httpService) GetThread(context.Context, uuid.UUID, uuid.UUID) (domain.Thread, error) {
	return domain.Thread{ID: uuid.New(), Version: 1}, service.err
}

// ListThreads lists threads.
func (service httpService) ListThreads(
	context.Context,
	uuid.UUID,
	port.ThreadFilter,
	pagination.Page,
) (pagination.Result[domain.Thread], error) {
	return pagination.Result[domain.Thread]{}, service.err
}

// UpdateThreadTitle updates thread title.
func (service httpService) UpdateThreadTitle(context.Context, port.UpdateThreadTitleCommand) (domain.Thread, error) {
	return domain.Thread{ID: uuid.New(), Version: 1}, service.err
}

// DeleteThread deletes one thread.
func (service httpService) DeleteThread(context.Context, port.DeleteThreadCommand) error {
	return service.err
}

// CreateReply creates a reply.
func (service httpService) CreateReply(context.Context, port.CreateReplyCommand) (domain.Post, error) {
	return domain.Post{ID: uuid.New(), Version: 1}, service.err
}

// ListPosts lists posts.
func (service httpService) ListPosts(context.Context, uuid.UUID, port.PostFilter, pagination.Page) (pagination.Result[domain.Post], error) {
	return pagination.Result[domain.Post]{}, service.err
}

// GetPost returns one post.
func (service httpService) GetPost(context.Context, uuid.UUID, uuid.UUID) (domain.Post, error) {
	return domain.Post{ID: uuid.New(), Version: 1}, service.err
}

// UpdatePost updates one post.
func (service httpService) UpdatePost(context.Context, port.UpdatePostCommand) (domain.Post, error) {
	return domain.Post{ID: uuid.New(), Version: 1}, service.err
}

// DeletePost deletes one post.
func (service httpService) DeletePost(context.Context, port.DeletePostCommand) error {
	return service.err
}

// ListPostRevisions lists revisions.
func (service httpService) ListPostRevisions(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	pagination.Page,
) (pagination.Result[domain.PostRevision], error) {
	return pagination.Result[domain.PostRevision]{}, service.err
}

// LikePost likes one post.
func (service httpService) LikePost(context.Context, port.LikePostCommand) (domain.PostLikeSummary, error) {
	return domain.PostLikeSummary{PostID: uuid.New(), LikeCount: 1, LikedByActor: true}, service.err
}

// UnlikePost unlikes one post.
func (service httpService) UnlikePost(context.Context, port.UnlikePostCommand) (domain.PostLikeSummary, error) {
	return domain.PostLikeSummary{PostID: uuid.New(), LikeCount: 0}, service.err
}

// ListLatestPosts lists latest posts.
func (service httpService) ListLatestPosts(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	pagination.Page,
) (pagination.Result[domain.LatestPostSummary], error) {
	return pagination.Result[domain.LatestPostSummary]{
		Items: []domain.LatestPostSummary{
			{ForumID: uuid.New(), ThreadID: uuid.New(), PostID: uuid.New(), AuthorUserID: uuid.New(), Sequence: 1, ThreadTitle: "Thread"},
		},
	}, service.err
}

// ListMostLikedPosts lists most-liked posts.
func (service httpService) ListMostLikedPosts(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	pagination.Page,
) (pagination.Result[domain.MostLikedPost], error) {
	return pagination.Result[domain.MostLikedPost]{
		Items: []domain.MostLikedPost{
			{
				ForumID:      uuid.New(),
				ThreadID:     uuid.New(),
				PostID:       uuid.New(),
				AuthorUserID: uuid.New(),
				Sequence:     1,
				ThreadTitle:  "Thread",
				LikeCount:    3,
			},
		},
	}, service.err
}

// MarkThreadRead marks a thread read.
func (service httpService) MarkThreadRead(context.Context, port.MarkThreadReadCommand) (domain.ThreadReadState, error) {
	return domain.ThreadReadState{
		ID:                   uuid.New(),
		UserID:               uuid.New(),
		ForumID:              uuid.New(),
		ThreadID:             uuid.New(),
		LastReadPostSequence: 1,
	}, service.err
}

// MarkForumRead marks a forum read.
func (service httpService) MarkForumRead(context.Context, port.MarkForumReadCommand) error {
	return service.err
}

// GetUnreadSummary returns unread counts.
func (service httpService) GetUnreadSummary(context.Context, uuid.UUID) (domain.UnreadSummary, error) {
	return domain.UnreadSummary{UserID: uuid.New(), UnreadThreadCount: 1}, service.err
}

// Search searches forum content.
func (service httpService) Search(context.Context, port.SearchCommand, pagination.Page) (pagination.Result[domain.SearchResult], error) {
	return pagination.Result[domain.SearchResult]{
		Items: []domain.SearchResult{{Type: "thread", ForumID: uuid.New(), ThreadID: uuid.New(), Title: "Thread"}},
	}, service.err
}

// VerifyStats reports stats drift.
func (service httpService) VerifyStats(context.Context) (domain.CounterDriftReport, error) {
	return domain.CounterDriftReport{}, service.err
}

// RebuildStats repairs stats drift.
func (service httpService) RebuildStats(context.Context) (domain.CounterDriftReport, error) {
	return domain.CounterDriftReport{Repaired: true}, service.err
}

// VerifyLikes reports like drift.
func (service httpService) VerifyLikes(context.Context) (domain.CounterDriftReport, error) {
	return domain.CounterDriftReport{}, service.err
}

// RebuildLikes repairs like drift.
func (service httpService) RebuildLikes(context.Context) (domain.CounterDriftReport, error) {
	return domain.CounterDriftReport{Repaired: true}, service.err
}

// FlushThreadViews flushes buffered thread views.
func (service httpService) FlushThreadViews(context.Context) (int64, error) {
	return 0, service.err
}

// ClearReadCache clears forum read caches.
func (service httpService) ClearReadCache(context.Context) error {
	return service.err
}
