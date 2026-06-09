package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// TestTreeAllowsAnonymous verifies the tree route accepts anonymous callers.
func TestTreeAllowsAnonymous(t *testing.T) {
	app := newTestApp(httpService{tree: domain.ForumTree{Categories: []domain.CategoryNode{}}})
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/forums/tree", nil)

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
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/forum-categories", bytes.NewBufferString(`{}`))
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
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/forum-categories", bytes.NewBufferString(`{"key":"official","name":"Official"}`))
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
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/forum-categories", bytes.NewBufferString(`{"key":"official","name":"Official"}`))
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
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/forums/tree", nil)
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
	req, _ := http.NewRequest(http.MethodPatch, "/api/v1/forums/"+uuid.NewString(), bytes.NewBufferString(`{}`))
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

// newTestApp creates a Fiber app with forum routes.
func newTestApp(service httpService) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(headers.Middleware())
	v1 := app.Group("/api/v1")
	Register(v1, Services{Forums: service})
	return app
}

// httpService is a forum service test double.
type httpService struct {
	tree domain.ForumTree
	err  error
}

// CreateCategory creates a category.
func (service httpService) CreateCategory(context.Context, port.CreateCategoryCommand) (domain.ForumCategory, error) {
	return domain.ForumCategory{ID: uuid.New(), Key: "official", Name: "Official", Status: domain.CategoryStatusActive, Version: 1}, service.err
}

// UpdateCategory updates a category.
func (service httpService) UpdateCategory(context.Context, port.UpdateCategoryCommand) (domain.ForumCategory, error) {
	return domain.ForumCategory{}, service.err
}

// GetCategory returns one category.
func (service httpService) GetCategory(context.Context, uuid.UUID) (domain.ForumCategory, error) {
	return domain.ForumCategory{}, service.err
}

// ListCategories lists categories.
func (service httpService) ListCategories(context.Context, port.CategoryFilter, pagination.Page) (pagination.Result[domain.ForumCategory], error) {
	return pagination.Result[domain.ForumCategory]{}, service.err
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
	return domain.Forum{}, service.err
}

// UpdateForum updates a forum.
func (service httpService) UpdateForum(context.Context, port.UpdateForumCommand) (domain.Forum, error) {
	return domain.Forum{}, service.err
}

// MoveForum moves a forum.
func (service httpService) MoveForum(context.Context, port.MoveForumCommand) (domain.Forum, error) {
	return domain.Forum{}, service.err
}

// GetForum returns one forum.
func (service httpService) GetForum(context.Context, uuid.UUID) (domain.Forum, error) {
	return domain.Forum{}, service.err
}

// ListForums lists forums.
func (service httpService) ListForums(context.Context, port.ForumFilter, pagination.Page) (pagination.Result[domain.Forum], error) {
	return pagination.Result[domain.Forum]{}, service.err
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
