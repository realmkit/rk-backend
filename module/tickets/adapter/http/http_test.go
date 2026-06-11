package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/tickets/domain"
	"github.com/niflaot/gamehub-go/module/tickets/port"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// TestCreateTicketRequiresIdempotency verifies retryable intake headers.
func TestCreateTicketRequiresIdempotency(t *testing.T) {
	app, service := newApp()
	service.ticket = validHTTPQueueTicket()
	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBufferString(`{}`))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(currentUserIDHeader, uuid.New().String())
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

// TestCreateTicketSuccess verifies intake DTO mapping and ETag response.
func TestCreateTicketSuccess(t *testing.T) {
	app, service := newApp()
	definitionID := uuid.New()
	service.ticket = validHTTPQueueTicket()
	body := `{"definition_id":"` + definitionID.String() + `","title":"Appeal","content_document_json":{"type":"doc"},"content_text":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewBufferString(body))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(currentUserIDHeader, uuid.New().String())
	req.Header.Set(headers.IdempotencyKey, "ticket-1")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if service.createCommand.DefinitionID != definitionID || service.createCommand.IdempotencyKey != "ticket-1" {
		t.Fatalf("command = %+v", service.createCommand)
	}
	if resp.Header.Get(headers.ETag) == "" {
		t.Fatalf("ETag header missing")
	}
}

// TestCloseTicketRequiresIfMatch verifies versioned workflow headers.
func TestCloseTicketRequiresIfMatch(t *testing.T) {
	app, _ := newApp()
	req := httptest.NewRequest(http.MethodPost, "/tickets/"+uuid.New().String()+"/close", bytes.NewBufferString(`{"reason":"done"}`))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(currentUserIDHeader, uuid.New().String())
	req.Header.Set(headers.IdempotencyKey, "close-1")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusPreconditionRequired {
		t.Fatalf("status = %d, want 428", resp.StatusCode)
	}
}

// TestCreateMessageAndEvidenceRoutes verifies conversation mapping.
func TestCreateMessageAndEvidenceRoutes(t *testing.T) {
	app, service := newApp()
	service.message = domain.Message{ID: uuid.New(), TicketID: uuid.New(), Version: 1}
	ticketID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/tickets/"+ticketID.String()+"/messages", bytes.NewBufferString(`{"content_document_json":{"type":"doc"},"content_text":"hello"}`))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(currentUserIDHeader, uuid.New().String())
	req.Header.Set(headers.IdempotencyKey, "msg-1")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("message status = %d, want 201", resp.StatusCode)
	}
	if service.messageCommand.TicketID != ticketID {
		t.Fatalf("message ticket = %s, want %s", service.messageCommand.TicketID, ticketID)
	}
	evidenceBody := `{"external_url":"https://example.com/proof","label":"proof"}`
	req = httptest.NewRequest(http.MethodPost, "/tickets/"+ticketID.String()+"/evidence", bytes.NewBufferString(evidenceBody))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(currentUserIDHeader, uuid.New().String())
	req.Header.Set(headers.IdempotencyKey, "evidence-1")
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("evidence app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("evidence status = %d, want 201", resp.StatusCode)
	}
	for _, path := range []string{
		"/tickets/" + ticketID.String() + "/messages",
		"/tickets/" + ticketID.String() + "/evidence",
	} {
		req = httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set(currentUserIDHeader, uuid.New().String())
		resp, err = app.Test(req)
		if err != nil {
			t.Fatalf("GET %s error = %v", path, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("GET %s status = %d, want 200", path, resp.StatusCode)
		}
	}
}

// TestTicketReadAndAppealRoutes verifies queue reads and appeal intake.
func TestTicketReadAndAppealRoutes(t *testing.T) {
	app, service := newApp()
	service.ticket = validHTTPQueueTicket()
	for _, path := range []string{"/tickets", "/tickets/" + uuid.New().String()} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set(currentUserIDHeader, uuid.New().String())
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("GET %s error = %v", path, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("GET %s status = %d, want 200", path, resp.StatusCode)
		}
	}
	definitionID := uuid.New()
	body := `{"definition_id":"` + definitionID.String() + `","title":"Appeal","content_document_json":{"type":"doc"},"content_text":"hello"}`
	punishmentID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/punishments/"+punishmentID.String()+"/appeals", bytes.NewBufferString(body))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(currentUserIDHeader, uuid.New().String())
	req.Header.Set(headers.IdempotencyKey, "appeal-route")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("appeal app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("appeal status = %d, want 201", resp.StatusCode)
	}
	if service.createCommand.PunishmentID == nil || *service.createCommand.PunishmentID != punishmentID {
		t.Fatalf("PunishmentID = %v, want %s", service.createCommand.PunishmentID, punishmentID)
	}
}

// TestDefinitionRoutes verifies definition CRUD adapter paths.
func TestDefinitionRoutes(t *testing.T) {
	app, _ := newApp()
	body := `{"key":"support","name":"Support","kind":"support","status":"active"}`
	req := httptest.NewRequest(http.MethodPost, "/ticket-definitions", bytes.NewBufferString(body))
	req.Header.Set(headers.ContentType, "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("create definition error = %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("create status = %d, want 201", resp.StatusCode)
	}
	id := uuid.New().String()
	for _, tt := range []struct {
		method string
		path   string
		status int
		body   string
	}{
		{method: http.MethodGet, path: "/ticket-definitions", status: fiber.StatusOK},
		{method: http.MethodGet, path: "/ticket-definitions/" + id, status: fiber.StatusOK},
		{method: http.MethodPatch, path: "/ticket-definitions/" + id, status: fiber.StatusOK, body: body},
		{method: http.MethodDelete, path: "/ticket-definitions/" + id, status: fiber.StatusNoContent},
	} {
		req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
		req.Header.Set(headers.ContentType, "application/json")
		if tt.method == http.MethodPatch || tt.method == http.MethodDelete {
			req.Header.Set(headers.IfMatch, `"1"`)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s error = %v", tt.method, tt.path, err)
		}
		if resp.StatusCode != tt.status {
			t.Fatalf("%s %s status = %d, want %d", tt.method, tt.path, resp.StatusCode, tt.status)
		}
	}
}

// TestActionAndOperationRoutes verifies workflow and operations adapters.
func TestActionAndOperationRoutes(t *testing.T) {
	app, service := newApp()
	service.ticket = validHTTPQueueTicket()
	ticketID := uuid.New().String()
	for _, path := range []string{
		"/tickets/" + ticketID + "/assign",
		"/tickets/" + ticketID + "/escalate",
		"/tickets/" + ticketID + "/close",
		"/tickets/" + ticketID + "/reopen",
		"/tickets/" + ticketID + "/appeal/accept",
		"/tickets/" + ticketID + "/appeal/reject",
	} {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(`{"reason":"ok"}`))
		req.Header.Set(headers.ContentType, "application/json")
		req.Header.Set(currentUserIDHeader, uuid.New().String())
		req.Header.Set(headers.IdempotencyKey, "workflow-"+path)
		req.Header.Set(headers.IfMatch, `"7"`)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s error = %v", path, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("%s status = %d, want 200", path, resp.StatusCode)
		}
	}
	for _, path := range []string{"/tickets/operations/stats/verify", "/tickets/operations/stats/rebuild"} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s error = %v", path, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("%s status = %d, want 200", path, resp.StatusCode)
		}
	}
}

// newApp creates a Fiber app with fake ticket services.
func newApp() (*fiber.App, *httpService) {
	service := &httpService{}
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	Register(app, Services{
		Definitions:  service,
		Tickets:      service,
		Conversation: service,
		Operations:   service,
	})
	return app, service
}

// httpService is a fake implementing all ticket HTTP service ports.
type httpService struct {
	ticket         domain.Ticket
	message        domain.Message
	createCommand  port.CreateTicketCommand
	messageCommand port.MessageCommand
}

// CreateDefinition returns the definition.
func (service *httpService) CreateDefinition(_ context.Context, definition domain.Definition) (domain.Definition, error) {
	return definition, nil
}

// UpdateDefinition returns the definition.
func (service *httpService) UpdateDefinition(_ context.Context, definition domain.Definition, _ uint64) (domain.Definition, error) {
	return definition, nil
}

// DeleteDefinition deletes nothing.
func (service *httpService) DeleteDefinition(context.Context, uuid.UUID, uint64) error { return nil }

// GetDefinition returns one definition.
func (service *httpService) GetDefinition(context.Context, uuid.UUID) (domain.Definition, error) {
	return domain.Definition{ID: uuid.New(), Key: "support", Name: "Support", Kind: domain.KindSupport, Version: 1}, nil
}

// ListDefinitions returns no definitions.
func (service *httpService) ListDefinitions(context.Context, port.DefinitionFilter, pagination.Page) (pagination.Result[domain.Definition], error) {
	return pagination.Result[domain.Definition]{}, nil
}

// CreateTicket records a create command.
func (service *httpService) CreateTicket(_ context.Context, command port.CreateTicketCommand) (domain.Ticket, error) {
	service.createCommand = command
	return service.ticket, nil
}

// GetTicket returns one ticket.
func (service *httpService) GetTicket(context.Context, uuid.UUID, uuid.UUID) (domain.Ticket, error) {
	return service.ticket, nil
}

// ListTickets returns no tickets.
func (service *httpService) ListTickets(context.Context, port.TicketFilter, pagination.Page) (pagination.Result[domain.Ticket], error) {
	return pagination.Result[domain.Ticket]{}, nil
}

// AssignTicket returns one ticket.
func (service *httpService) AssignTicket(context.Context, port.StaffCommand) (domain.Ticket, error) {
	return service.ticket, nil
}

// EscalateTicket returns one ticket.
func (service *httpService) EscalateTicket(context.Context, port.StaffCommand) (domain.Ticket, error) {
	return service.ticket, nil
}

// CloseTicket returns one ticket.
func (service *httpService) CloseTicket(context.Context, port.StaffCommand) (domain.Ticket, error) {
	return service.ticket, nil
}

// ReopenTicket returns one ticket.
func (service *httpService) ReopenTicket(context.Context, port.StaffCommand) (domain.Ticket, error) {
	return service.ticket, nil
}

// AcceptAppeal returns one ticket.
func (service *httpService) AcceptAppeal(context.Context, port.AppealDecisionCommand) (domain.Ticket, error) {
	return service.ticket, nil
}

// RejectAppeal returns one ticket.
func (service *httpService) RejectAppeal(context.Context, port.AppealDecisionCommand) (domain.Ticket, error) {
	return service.ticket, nil
}

// CreateMessage records a message command.
func (service *httpService) CreateMessage(_ context.Context, command port.MessageCommand) (domain.Message, error) {
	service.messageCommand = command
	return service.message, nil
}

// ListMessages returns no messages.
func (service *httpService) ListMessages(context.Context, uuid.UUID, uuid.UUID, bool, pagination.Page) (pagination.Result[domain.Message], error) {
	return pagination.Result[domain.Message]{}, nil
}

// AddEvidence returns evidence.
func (service *httpService) AddEvidence(_ context.Context, command port.EvidenceCommand) (domain.Evidence, error) {
	return domain.Evidence{ID: uuid.New(), TicketID: command.TicketID, CreatedAt: time.Now().UTC()}, nil
}

// ListEvidence returns no evidence.
func (service *httpService) ListEvidence(context.Context, uuid.UUID, uuid.UUID, bool) ([]domain.Evidence, error) {
	return nil, nil
}

// VerifyStats returns an empty report.
func (service *httpService) VerifyStats(context.Context) (domain.DriftReport, error) {
	return domain.DriftReport{}, nil
}

// RebuildStats returns a repaired report.
func (service *httpService) RebuildStats(context.Context) (domain.DriftReport, error) {
	return domain.DriftReport{Repaired: true}, nil
}

// DetectSLABreaches returns no changes.
func (service *httpService) DetectSLABreaches(context.Context) (int64, error) { return 0, nil }

// CloseStaleTickets returns no changes.
func (service *httpService) CloseStaleTickets(context.Context) (int64, error) { return 0, nil }

// ClearCache clears nothing.
func (service *httpService) ClearCache(context.Context) error { return nil }

// validHTTPQueueTicket returns a valid response ticket.
func validHTTPQueueTicket() domain.Ticket {
	return domain.Ticket{
		ID:           uuid.New(),
		DefinitionID: uuid.New(),
		Title:        "Ticket",
		Kind:         domain.KindSupport,
		Status:       domain.StatusOpen,
		Priority:     domain.PriorityNormal,
		OpenedAt:     time.Now().UTC(),
		Version:      7,
	}
}
