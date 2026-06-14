// Package forums_e2e verifies forum journeys through the real server.
package forums_e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	forumshttp "github.com/realmkit/rk-backend/module/forums/adapter/http"
	forumspostgres "github.com/realmkit/rk-backend/module/forums/adapter/postgres"
	forumsredis "github.com/realmkit/rk-backend/module/forums/adapter/redis"
	forumsapplication "github.com/realmkit/rk-backend/module/forums/application"
	forumsdomain "github.com/realmkit/rk-backend/module/forums/domain"
	groupspostgres "github.com/realmkit/rk-backend/module/groups/adapter/postgres"
	groupsapplication "github.com/realmkit/rk-backend/module/groups/application"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/openapi"
	eventtesting "github.com/realmkit/rk-backend/pkg/events/testing"
	"github.com/realmkit/rk-backend/pkg/server"
	"github.com/realmkit/rk-backend/pkg/transaction"
	goredis "github.com/redis/go-redis/v9"
)

type forumsFixture struct {
	ecosystem    *harness.Ecosystem
	service      forumsapplication.Service
	groups       groupsapplication.Service
	assets       *forumAssets
	restrictions *forumRestrictions
	events       *eventtesting.PublisherRecorder
	manager      uuid.UUID
	member       uuid.UUID
	other        uuid.UUID
}

func newForumsFixture(t *testing.T) forumsFixture {
	t.Helper()
	database := harness.NewSQLiteDatabase(t)
	events := &eventtesting.PublisherRecorder{}
	redisServer := miniredis.RunT(t)
	redisClient := goredis.NewClient(&goredis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close(); redisServer.Close() })
	groups := groupsapplication.NewService(
		groupspostgres.NewGroupRepository(database.Store),
		groupspostgres.NewMembershipRepository(database.Store),
		groupspostgres.NewPermissionRepository(database.Store),
	)
	assets := &forumAssets{known: map[uuid.UUID]bool{}}
	restrictions := &forumRestrictions{blocked: map[uuid.UUID]map[string]bool{}}
	service := forumsapplication.NewService(forumsapplication.Dependencies{
		Categories:   forumspostgres.NewCategoryRepository(database.Store),
		Forums:       forumspostgres.NewForumRepository(database.Store),
		Threads:      forumspostgres.NewThreadRepository(database.Store),
		Posts:        forumspostgres.NewPostRepository(database.Store),
		Interactions: forumspostgres.NewInteractionRepository(database.Store),
		Operations:   forumspostgres.NewOperationsRepository(database.Store),
		Authorizer:   forumspostgres.NewVisibilityAuthorizer(database.Store),
		Assets:       assets,
		Restrictions: restrictions,
		Cache:        forumsredis.NewTreeCache(redisClient),
		Transactions: transaction.New(database.DB),
		Events:       events,
	})
	ecosystem := harness.New(t,
		harness.WithDevelopment(true),
		harness.WithDatabase(database),
		harness.WithServerOptions(
			server.WithAuth(auth.Config{DevelopmentBypass: true}, harness.DevProvisioner{}),
			server.WithForums(forumshttp.Services{
				Structure: service,
				Content:   service, Interaction: service,
				Operations: service, Admin: service,
			}),
		),
	)
	fixture := forumsFixture{
		ecosystem: ecosystem, service: service, groups: groups,
		assets: assets, restrictions: restrictions, events: events,
		manager: uuid.New(), member: uuid.New(), other: uuid.New(),
	}
	fixture.grant(t, forumsdomain.RootForumObjectID(), groupsdomain.RelationManager, fixture.manager)
	return fixture
}

type forumAssets struct{ known map[uuid.UUID]bool }

func (assets *forumAssets) AssetExists(_ context.Context, id uuid.UUID) (bool, error) {
	return assets.known[id], nil
}

type forumRestrictions struct{ blocked map[uuid.UUID]map[string]bool }

func (checker *forumRestrictions) Restricted(_ context.Context, userID uuid.UUID, key string) (bool, error) {
	return checker.blocked[userID][key], nil
}

func (fixture forumsFixture) do(t *testing.T, request *http.Request) *http.Response {
	t.Helper()
	return fixture.ecosystem.Test(t, request)
}

func (fixture forumsFixture) grant(t *testing.T, objectID uuid.UUID, relation groupsdomain.Relation, user uuid.UUID) {
	t.Helper()
	_, err := fixture.groups.CreatePermissionGrant(context.Background(), groupsport.CreatePermissionGrantCommand{
		Grant: groupsdomain.PermissionGrant{
			SubjectType: groupsdomain.SubjectUser,
			SubjectID:   user,
			Action:      forumActionForRelation(relation),
			ScopeType:   groupsdomain.ObjectForum,
			ScopeID:     objectID,
		},
	})
	if err != nil {
		t.Fatalf("CreatePermissionGrant(%s) error = %v", relation, err)
	}
	if err := fixture.service.ClearReadCache(context.Background()); err != nil {
		t.Fatalf("ClearReadCache() error = %v", err)
	}
}

func forumActionForRelation(relation groupsdomain.Relation) groupsdomain.Action {
	switch relation {
	case groupsdomain.RelationViewer:
		return groupsdomain.PermissionForumsView
	case groupsdomain.RelationCreator:
		return groupsdomain.PermissionForumsCreateThread
	case groupsdomain.RelationReplyer:
		return groupsdomain.PermissionForumsReply
	case groupsdomain.RelationLiker:
		return groupsdomain.PermissionForumsLikePosts
	case groupsdomain.RelationModerator:
		return groupsdomain.PermissionForumsManageThreads
	default:
		return groupsdomain.PermissionForumsManageForum
	}
}

func forumRequest(method string, path string, body string, opts ...func(*http.Request)) *http.Request {
	request := harness.JSONRequest(method, path, body)
	for _, opt := range opts {
		opt(request)
	}
	return request
}

func forumUser(userID uuid.UUID) func(*http.Request) {
	return func(request *http.Request) { request.Header.Set(auth.DevUserIDHeader, userID.String()) }
}

func forumIdempotency(key string) func(*http.Request) {
	return func(request *http.Request) { request.Header.Set(headers.IdempotencyKey, key) }
}

func forumIfMatch(version uint64) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set(headers.IfMatch, `"`+strconv.FormatUint(version, 10)+`"`)
	}
}

func decodeForumObject(t *testing.T, response *http.Response) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return payload
}

func decodeForumList(t *testing.T, response *http.Response) []any {
	t.Helper()
	return decodeForumObject(t, response)["items"].([]any)
}

func assertForumStatus(t *testing.T, response *http.Response, want int) {
	t.Helper()
	if response.StatusCode != want {
		t.Fatalf("status = %d, want %d body = %q", response.StatusCode, want, harness.ResponseBody(t, response))
	}
}

func forumID(t *testing.T, payload map[string]any, field string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(payload[field].(string))
	if err != nil {
		t.Fatalf("Parse(%s) error = %v", field, err)
	}
	return id
}

func forumVersion(payload map[string]any) uint64 { return uint64(payload["version"].(float64)) }

func assertForumOpenAPIRoute(t *testing.T, method string, path string) {
	t.Helper()
	ok, err := openapi.OperationExists(method, path)
	if err != nil {
		t.Fatalf("OperationExists() error = %v", err)
	}
	if !ok {
		t.Fatalf("%s %s missing OpenAPI operation", method, path)
	}
}

func threadBody(title string) string {
	return `{"title":"` + title + `","slug":"` + slug(title) +
		`","content_document_json":{"type":"doc","content":[{"type":"text","text":"` + title + `"}]}}`
}

func postBody(text string) string {
	return `{"content_document_json":{"type":"doc","content":[{"type":"text","text":"` + text + `"}]}}`
}

func slug(value string) string {
	out := ""
	for _, r := range value {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out += string(r)
		} else if out != "" && out[len(out)-1] != '-' {
			out += "-"
		}
	}
	if out == "" {
		return "forum-e2e"
	}
	return out
}
