package http

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/events/application"
	"github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// TestEventHTTPProblemMappings covers route-level error translation.
func TestEventHTTPProblemMappings(t *testing.T) {
	validationErr := domain.ErrorIfInvalid([]domain.Violation{
		{Field: "scope", Message: "is required"},
	})
	cases := []struct {
		name   string
		err    error
		method string
		path   string
		status int
	}{
		{"invalid user", nil, http.MethodGet, "/events", fiber.StatusBadRequest},
		{"invalid pagination", nil, http.MethodGet, "/events?page_size=-1", fiber.StatusBadRequest},
		{"not found", port.ErrNotFound, http.MethodGet, "/events/" + uuid.NewString(), fiber.StatusNotFound},
		{"forbidden", port.ErrForbidden, http.MethodPost, "/events/" + uuid.NewString() + "/cancel", fiber.StatusForbidden},
		{"validation", validationErr, http.MethodGet, "/events", fiber.StatusUnprocessableEntity},
	}

	for _, item := range cases {
		app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
		useTestPrincipal(app)
		Register(app, Services{Events: eventServiceWithError(item.err)})
		req := newRequest(t, item.method, item.path)
		req.Header.Set(currentUserIDHeader, uuid.NewString())
		if item.name == "invalid user" {
			req.Header.Set(currentUserIDHeader, "not-a-uuid")
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s error = %v", item.name, err)
		}
		if resp.StatusCode != item.status {
			t.Fatalf("%s status = %d, want %d", item.name, resp.StatusCode, item.status)
		}
	}
}

// eventServiceWithError creates an event service backed by an erroring repository.
func eventServiceWithError(err error) application.Service {
	return application.NewService(application.Dependencies{Repository: eventRepoFake{err: err}})
}

type eventRepoFake struct {
	err error
}

func (repo eventRepoFake) Publish(
	context.Context,
	domain.Draft,
	time.Time,
) (domain.Event, error) {
	return domain.Event{}, repo.err
}

func (repo eventRepoFake) Get(context.Context, uuid.UUID) (domain.Event, error) {
	if repo.err != nil {
		return domain.Event{}, repo.err
	}
	return domain.Event{ID: uuid.New()}, nil
}

func (repo eventRepoFake) List(
	context.Context,
	port.ListFilter,
	pagination.Page,
) (pagination.Result[domain.Event], error) {
	if repo.err != nil {
		return pagination.Result[domain.Event]{}, repo.err
	}
	return pagination.Result[domain.Event]{}, nil
}

func (repo eventRepoFake) Claim(
	context.Context,
	string,
	int,
	time.Time,
	time.Time,
) ([]domain.Event, error) {
	return nil, repo.err
}

func (repo eventRepoFake) MarkProcessed(context.Context, uuid.UUID, time.Time) error {
	return repo.err
}

func (repo eventRepoFake) MarkFailed(context.Context, uuid.UUID, string, time.Time, time.Time) error {
	return repo.err
}

func (repo eventRepoFake) MarkDead(context.Context, uuid.UUID, string, time.Time) error {
	return repo.err
}

func (repo eventRepoFake) Replay(context.Context, uuid.UUID, time.Time) error {
	return repo.err
}

func (repo eventRepoFake) Cancel(context.Context, uuid.UUID, time.Time) error {
	return repo.err
}
