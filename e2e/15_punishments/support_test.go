// Package punishments_e2e verifies punishment journeys through the real server.
package punishments_e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	punishmentshttp "github.com/realmkit/rk-backend/module/punishments/adapter/http"
	punishmentspostgres "github.com/realmkit/rk-backend/module/punishments/adapter/postgres"
	punishmentsredis "github.com/realmkit/rk-backend/module/punishments/adapter/redis"
	punishmentsapplication "github.com/realmkit/rk-backend/module/punishments/application"
	punishmentsdomain "github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/openapi"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	eventtesting "github.com/realmkit/rk-backend/pkg/events/testing"
	"github.com/realmkit/rk-backend/pkg/server"
	"github.com/realmkit/rk-backend/pkg/transaction"
	goredis "github.com/redis/go-redis/v9"
)

// punishmentsFixture contains punishment e2e wiring.
type punishmentsFixture struct {
	ecosystem *harness.Ecosystem
	service   punishmentsapplication.Service
	events    *eventtesting.PublisherRecorder
}

// newPunishmentsFixture starts a server with punishment routes.
func newPunishmentsFixture(t *testing.T) punishmentsFixture {
	t.Helper()
	database := harness.NewSQLiteDatabase(t)
	events := &eventtesting.PublisherRecorder{}
	redisServer := miniredis.RunT(t)
	redisClient := goredis.NewClient(&goredis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() {
		if err := redisClient.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		redisServer.Close()
	})
	service := punishmentsapplication.NewService(punishmentsapplication.Dependencies{
		Definitions:  punishmentspostgres.NewDefinitionRepository(database.Store),
		Cases:        punishmentspostgres.NewCaseRepository(database.Store),
		Cache:        punishmentsredis.NewCache(redisClient),
		Transactions: transaction.New(database.DB),
		Events:       events,
	})
	ecosystem := harness.New(
		t,
		harness.WithDevelopment(true),
		harness.WithDatabase(database),
		harness.WithServerOptions(
			server.WithAuth(auth.Config{DevelopmentBypass: true}, harness.DevProvisioner{}),
			server.WithPunishments(punishmentshttp.Services{Punishments: service, Checker: harness.AllowChecker{}}),
		),
	)
	return punishmentsFixture{ecosystem: ecosystem, service: service, events: events}
}

// do sends a request through the fixture server.
func (fixture punishmentsFixture) do(t *testing.T, request *http.Request) *http.Response {
	t.Helper()
	return fixture.ecosystem.Test(t, request)
}

// createDefinition stores one active definition through HTTP.
func (fixture punishmentsFixture) createDefinition(
	t *testing.T,
	actor uuid.UUID,
	key string,
	actions ...punishmentsdomain.ActionType,
) map[string]any {
	t.Helper()
	response := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodPost, "/punishment-definitions", definitionBody(key, actions...)),
			withPunishmentUser(actor),
			withPunishmentIdempotency("definition-"+key),
		),
	)
	assertPunishmentStatus(t, response, fiber.StatusCreated)
	return decodePunishmentObject(t, response)
}

// issuePunishment creates an active punishment through HTTP.
func (fixture punishmentsFixture) issuePunishment(
	t *testing.T,
	actor uuid.UUID,
	definitionID uuid.UUID,
	targetID uuid.UUID,
	key string,
	expiresAt *time.Time,
) map[string]any {
	t.Helper()
	body := issueBody(definitionID, targetID, expiresAt)
	response := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodPost, "/punishments", body),
			withPunishmentUser(actor),
			withPunishmentIdempotency("issue-"+key),
		),
	)
	assertPunishmentStatus(t, response, fiber.StatusCreated)
	return decodePunishmentObject(t, response)
}

// definitionBody returns a valid punishment definition JSON body.
func definitionBody(key string, actions ...punishmentsdomain.ActionType) string {
	if len(actions) == 0 {
		actions = []punishmentsdomain.ActionType{punishmentsdomain.ActionForumsReply}
	}
	rawActions := ""
	for index, action := range actions {
		if index > 0 {
			rawActions += ","
		}
		rawActions += `{"target_system":"realmkit","action_type":"` + string(action) +
			`","configuration_json":{},"display_order":` +
			strconv.Itoa(index+1) + `,"status":"active"}`
	}
	return `{"key":"` + key + `","name":"` + key + `","color":"#ff5555",` +
		`"severity":10,"status":"active","allow_permanent":true,` +
		`"requires_reason":true,"actions":[` + rawActions + `]}`
}

// issueBody returns a valid punishment issue body.
func issueBody(definitionID uuid.UUID, targetID uuid.UUID, expiresAt *time.Time) string {
	expires := "null"
	if expiresAt != nil {
		expires = `"` + expiresAt.Format(time.RFC3339Nano) + `"`
	}
	return `{"definition_id":"` + definitionID.String() + `","target_user_id":"` +
		targetID.String() + `","issuer_type":"user","reason":"E2E moderation",` +
		`"source":"e2e","expires_at":` + expires + `}`
}

// withPunishmentUser adds the current-user header.
func withPunishmentUser(userID uuid.UUID) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set(auth.DevUserIDHeader, userID.String())
	}
}

// withPunishmentIdempotency adds an idempotency key.
func withPunishmentIdempotency(key string) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set(headers.IdempotencyKey, key)
	}
}

// withPunishmentIfMatch adds an If-Match header.
func withPunishmentIfMatch(version uint64) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set(headers.IfMatch, `"`+strconv.FormatUint(version, 10)+`"`)
	}
}

// configureRequest applies request mutations.
func configureRequest(request *http.Request, configs ...func(*http.Request)) *http.Request {
	for _, config := range configs {
		config(request)
	}
	return request
}

// decodePunishmentObject decodes one JSON object.
func decodePunishmentObject(t *testing.T, response *http.Response) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return payload
}

// assertPunishmentStatus verifies response status.
func assertPunishmentStatus(t *testing.T, response *http.Response, want int) {
	t.Helper()
	if response.StatusCode != want {
		t.Fatalf("StatusCode = %d, want %d body = %q", response.StatusCode, want, harness.ResponseBody(t, response))
	}
}

// idFrom extracts an ID field.
func idFrom(t *testing.T, payload map[string]any, field string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(payload[field].(string))
	if err != nil {
		t.Fatalf("Parse(%s) error = %v", field, err)
	}
	return id
}

// versionFrom extracts the root version.
func versionFrom(payload map[string]any) uint64 {
	return uint64(payload["version"].(float64))
}

// assertPunishmentOpenAPIRoute verifies an OpenAPI operation exists.
func assertPunishmentOpenAPIRoute(t *testing.T, method string, path string) {
	t.Helper()
	ok, err := openapi.OperationExists(method, path)
	if err != nil {
		t.Fatalf("OperationExists() error = %v", err)
	}
	if !ok {
		t.Fatalf("%s %s missing OpenAPI operation", method, path)
	}
}

// assertPunishmentEvent verifies that an event key was published.
func assertPunishmentEvent(t *testing.T, fixture punishmentsFixture, key eventdomain.EventKey) {
	t.Helper()
	for _, draft := range fixture.events.Drafts() {
		if draft.Key == key && draft.Producer == eventdomain.ProducerPunishments {
			return
		}
	}
	t.Fatalf("event %s was not published", key)
}

// compile-time guard for context import used by service-level operations.
var _ = context.Background
