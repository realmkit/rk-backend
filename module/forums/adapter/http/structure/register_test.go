package structure

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

// TestStructureRoutesExerciseHandlers verifies structure route wiring.
func TestStructureRoutesExerciseHandlers(t *testing.T) {
	app := fiber.New()
	Register(app, Services{
		Structure: structureService{},
		Admin:     adminService{},
	})
	userID := uuid.NewString()
	for _, route := range structureRoutes() {
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

// structureRoute describes one route test case.
type structureRoute struct {
	method     string
	path       string
	body       string
	status     int
	idempotent bool
	versioned  bool
}

// structureRoutes returns route cases for structure handlers.
func structureRoutes() []structureRoute {
	categoryID := uuid.NewString()
	forumID := uuid.NewString()
	categoryBody := `{"key":"official","name":"Official"}`
	forumBody := `{"category_id":"` + uuid.NewString() + `","kind":"discussion","key":"news","slug":"news","name":"News"}`
	settingsBody := `{"thread_visibility_mode":"all_threads","max_sticky_threads":3,"default_thread_status":"open"}`
	permissionsBody := `{"viewers":[{"subject_type":"public"}]}`
	return []structureRoute{
		{method: http.MethodGet, path: "/forums/tree", status: fiber.StatusOK},
		{method: http.MethodPost, path: "/forum-categories", body: categoryBody, status: fiber.StatusCreated, idempotent: true},
		{method: http.MethodGet, path: "/forum-categories", status: fiber.StatusOK},
		{method: http.MethodGet, path: "/forum-categories/" + categoryID, status: fiber.StatusOK},
		{method: http.MethodPatch, path: "/forum-categories/" + categoryID, body: categoryBody, status: fiber.StatusOK, idempotent: true, versioned: true},
		{method: http.MethodDelete, path: "/forum-categories/" + categoryID, status: fiber.StatusNoContent, idempotent: true, versioned: true},
		{method: http.MethodPost, path: "/forum-categories/reorder", body: `{"items":[]}`, status: fiber.StatusNoContent, idempotent: true},
		{method: http.MethodPost, path: "/forums", body: forumBody, status: fiber.StatusCreated, idempotent: true},
		{method: http.MethodGet, path: "/forums", status: fiber.StatusOK},
		{method: http.MethodGet, path: "/forums/" + forumID + "/settings", status: fiber.StatusOK},
		{method: http.MethodPatch, path: "/forums/" + forumID + "/settings", body: settingsBody, status: fiber.StatusOK, idempotent: true, versioned: true},
		{method: http.MethodGet, path: "/forums/" + forumID + "/permissions", status: fiber.StatusOK},
		{method: http.MethodPut, path: "/forums/" + forumID + "/permissions", body: permissionsBody, status: fiber.StatusNoContent, idempotent: true},
		{method: http.MethodPost, path: "/forums/" + forumID + "/permissions/simulate", body: `{"permission":"forums.view","object_type":"forum"}`, status: fiber.StatusOK},
		{method: http.MethodGet, path: "/forums/" + forumID, status: fiber.StatusOK},
		{method: http.MethodPatch, path: "/forums/" + forumID, body: forumBody, status: fiber.StatusOK, idempotent: true, versioned: true},
		{method: http.MethodDelete, path: "/forums/" + forumID, status: fiber.StatusNoContent, idempotent: true, versioned: true},
		{method: http.MethodPost, path: "/forums/" + forumID + "/move", body: `{"category_id":"` + uuid.NewString() + `"}`, status: fiber.StatusOK, idempotent: true, versioned: true},
		{method: http.MethodPost, path: "/forums/reorder", body: `{"items":[]}`, status: fiber.StatusNoContent, idempotent: true},
	}
}

// structureService is a test structure service.
type structureService struct{}

// CreateCategory creates a test category.
func (structureService) CreateCategory(context.Context, port.CreateCategoryCommand) (domain.ForumCategory, error) {
	return domain.ForumCategory{ID: uuid.New(), Name: "Official", Version: 1}, nil
}

// UpdateCategory updates a test category.
func (structureService) UpdateCategory(context.Context, port.UpdateCategoryCommand) (domain.ForumCategory, error) {
	return domain.ForumCategory{ID: uuid.New(), Name: "Official", Version: 2}, nil
}

// GetCategory returns a test category.
func (structureService) GetCategory(context.Context, uuid.UUID) (domain.ForumCategory, error) {
	return domain.ForumCategory{ID: uuid.New(), Version: 1}, nil
}

// ListCategories returns a test category page.
func (structureService) ListCategories(context.Context, port.CategoryFilter, pagination.Page) (pagination.Result[domain.ForumCategory], error) {
	return pagination.Result[domain.ForumCategory]{Items: []domain.ForumCategory{}}, nil
}

// DeleteCategory deletes a test category.
func (structureService) DeleteCategory(context.Context, port.DeleteCategoryCommand) error { return nil }

// ReorderCategories reorders test categories.
func (structureService) ReorderCategories(context.Context, port.ReorderCategoriesCommand) error {
	return nil
}

// CreateForum creates a test forum.
func (structureService) CreateForum(context.Context, port.CreateForumCommand) (domain.Forum, error) {
	return domain.Forum{ID: uuid.New(), Version: 1}, nil
}

// UpdateForum updates a test forum.
func (structureService) UpdateForum(context.Context, port.UpdateForumCommand) (domain.Forum, error) {
	return domain.Forum{ID: uuid.New(), Version: 2}, nil
}

// MoveForum moves a test forum.
func (structureService) MoveForum(context.Context, port.MoveForumCommand) (domain.Forum, error) {
	return domain.Forum{ID: uuid.New(), Version: 2}, nil
}

// GetForum returns a test forum.
func (structureService) GetForum(context.Context, uuid.UUID) (domain.Forum, error) {
	return domain.Forum{ID: uuid.New(), Version: 1}, nil
}

// ListForums returns a test forum page.
func (structureService) ListForums(context.Context, port.ForumFilter, pagination.Page) (pagination.Result[domain.Forum], error) {
	return pagination.Result[domain.Forum]{Items: []domain.Forum{}}, nil
}

// DeleteForum deletes a test forum.
func (structureService) DeleteForum(context.Context, port.DeleteForumCommand) error { return nil }

// ReorderForums reorders test forums.
func (structureService) ReorderForums(context.Context, port.ReorderForumsCommand) error { return nil }

// Tree returns a test tree.
func (structureService) Tree(context.Context, uuid.UUID) (domain.ForumTree, error) {
	return domain.ForumTree{Categories: []domain.CategoryNode{}}, nil
}

// adminService is a test admin service.
type adminService struct{}

// GetForumSettings returns test settings.
func (adminService) GetForumSettings(context.Context, uuid.UUID, uuid.UUID) (domain.ForumSettings, error) {
	return domain.ForumSettings{ForumID: uuid.New(), Version: 1}, nil
}

// UpdateForumSettings updates test settings.
func (adminService) UpdateForumSettings(context.Context, port.UpdateForumSettingsCommand) (domain.ForumSettings, error) {
	return domain.ForumSettings{ForumID: uuid.New(), Version: 2}, nil
}

// GetForumPermissionSettings returns test permission settings.
func (adminService) GetForumPermissionSettings(context.Context, uuid.UUID, uuid.UUID) (domain.ForumPermissionSettings, error) {
	return domain.ForumPermissionSettings{}, nil
}

// UpdateForumPermissionSettings updates test permission settings.
func (adminService) UpdateForumPermissionSettings(context.Context, port.UpdateForumPermissionSettingsCommand) error {
	return nil
}

// SimulateForumPermission returns a test simulation.
func (adminService) SimulateForumPermission(context.Context, port.SimulateForumPermissionCommand) (domain.ForumPermissionSimulationResult, error) {
	return domain.ForumPermissionSimulationResult{Allowed: true}, nil
}
