package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/assets/domain"
	"github.com/realmkit/rk-backend/module/assets/port"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// TestCreateUploadIntentRequiresIdempotency verifies upload intent idempotency headers.
func TestCreateUploadIntentRequiresIdempotency(t *testing.T) {
	app := testApp(&httpService{})
	req := httptestRequest(http.MethodPost, "/assets/upload-intents", `{"namespace":"community"}`)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

// TestCreateUploadIntentReturnsCreated verifies successful upload intent response.
func TestCreateUploadIntentReturnsCreated(t *testing.T) {
	service := &httpService{asset: testHTTPAsset()}
	app := testApp(service)
	req := httptestRequest(
		http.MethodPost,
		"/assets/upload-intents",
		`{"namespace":"community","path":"brand","filename":"logo.png","visibility":"public","content_type":"image/png","size_bytes":512}`,
	)
	req.Header.Set(headers.IdempotencyKey, "intent-key")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated || resp.Header.Get(headers.ETag) == "" {
		t.Fatalf("status=%d etag=%q, want created with etag", resp.StatusCode, resp.Header.Get(headers.ETag))
	}
	if service.created == 0 {
		t.Fatalf("created = 0, want service call")
	}
	if service.createCommand.CreatedByUserID == nil {
		t.Fatalf("CreatedByUserID missing")
	}
}

// TestUpdateAssetRequiresIfMatch verifies optimistic concurrency header is required.
func TestUpdateAssetRequiresIfMatch(t *testing.T) {
	app := testApp(&httpService{asset: testHTTPAsset()})
	req := httptestRequest(
		http.MethodPatch,
		"/assets/"+testHTTPAsset().ID.String(),
		`{"namespace":"community","display_name":"Logo","path":"brand","visibility":"public"}`,
	)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusPreconditionRequired {
		t.Fatalf("StatusCode = %d, want %d", resp.StatusCode, fiber.StatusPreconditionRequired)
	}
}

// TestGetAssetMapsNotFound verifies service not found errors become problem responses.
func TestGetAssetMapsNotFound(t *testing.T) {
	app := testApp(&httpService{err: port.ErrNotFound})
	req := httptestRequest(http.MethodGet, "/assets/"+uuid.NewString(), "")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("StatusCode = %d, want %d", resp.StatusCode, fiber.StatusNotFound)
	}
}

// TestListFoldersReturnsFolders verifies virtual folder responses.
func TestListFoldersReturnsFolders(t *testing.T) {
	app := testApp(&httpService{folders: []string{"brand"}})
	req := httptestRequest(http.MethodGet, "/assets/folders?namespace=community", "")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK || !strings.Contains(string(body), "brand") {
		t.Fatalf("status=%d body=%s, want folders", resp.StatusCode, body)
	}
}

// TestListNamespacesReturnsNamespaces verifies namespace responses.
func TestListNamespacesReturnsNamespaces(t *testing.T) {
	app := testApp(&httpService{namespaces: []string{"community"}})
	req := httptestRequest(http.MethodGet, "/assets/namespaces", "")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK || !strings.Contains(string(body), "community") {
		t.Fatalf("status=%d body=%s, want namespaces", resp.StatusCode, body)
	}
}

// TestAssetRoutesExerciseReadAndMutationPaths verifies remaining asset routes.
func TestAssetRoutesExerciseReadAndMutationPaths(t *testing.T) {
	asset := testHTTPAsset()
	app := testApp(&httpService{asset: asset, folders: []string{"brand"}})
	cases := []struct {
		name   string
		method string
		path   string
		body   string
		status int
		header map[string]string
	}{
		{
			name:   "complete",
			method: http.MethodPost,
			path:   "/assets/" + asset.ID.String() + "/complete",
			status: fiber.StatusOK,
			header: map[string]string{headers.IdempotencyKey: "complete-key"},
		},
		{name: "get", method: http.MethodGet, path: "/assets/" + asset.ID.String(), status: fiber.StatusOK},
		{name: "url", method: http.MethodGet, path: "/assets/" + asset.ID.String() + "/url", status: fiber.StatusOK},
		{
			name:   "list",
			method: http.MethodGet,
			path:   "/assets?namespace=community&path=brand&path_prefix=brand&status=available&page_size=10",
			status: fiber.StatusOK,
		},
		{
			name:   "update",
			method: http.MethodPatch,
			path:   "/assets/" + asset.ID.String(),
			body:   `{"namespace":"community","display_name":"Logo","path":"brand","visibility":"public"}`,
			status: fiber.StatusOK,
			header: map[string]string{headers.IfMatch: `"1"`},
		},
		{
			name:   "delete",
			method: http.MethodDelete,
			path:   "/assets/" + asset.ID.String(),
			status: fiber.StatusNoContent,
			header: map[string]string{headers.IfMatch: `"1"`},
		},
	}
	for _, tt := range cases {
		req := httptestRequest(tt.method, tt.path, tt.body)
		for key, value := range tt.header {
			req.Header.Set(key, value)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s Test() error = %v", tt.name, err)
		}
		if resp.StatusCode != tt.status {
			t.Fatalf("%s StatusCode = %d, want %d", tt.name, resp.StatusCode, tt.status)
		}
	}
}

// TestPrivateAssetRoutesRequireUser verifies asset metadata and signed URLs are not anonymous.
func TestPrivateAssetRoutesRequireUser(t *testing.T) {
	assetID := testHTTPAsset().ID.String()
	app := testApp(&httpService{asset: testHTTPAsset()})
	for _, req := range []*http.Request{
		rawRequest(t, http.MethodPost, "/assets/upload-intents", `{}`),
		rawRequest(t, http.MethodPost, "/assets/"+assetID+"/complete", ``),
		rawRequest(t, http.MethodGet, "/assets/"+assetID, ``),
		rawRequest(t, http.MethodGet, "/assets/"+assetID+"/url", ``),
		rawRequest(t, http.MethodGet, "/assets", ``),
		rawRequest(t, http.MethodGet, "/assets/namespaces", ``),
		rawRequest(t, http.MethodGet, "/assets/folders", ``),
		rawRequest(t, http.MethodPatch, "/assets/"+assetID, `{}`),
		rawRequest(t, http.MethodDelete, "/assets/"+assetID, ``),
	} {
		req.Header.Set(headers.ContentType, "application/json")
		req.Header.Set(headers.IdempotencyKey, "asset-security")
		req.Header.Set(headers.IfMatch, `"1"`)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s error = %v", req.Method, req.URL.Path, err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Fatalf("%s %s status = %d, want 401", req.Method, req.URL.Path, resp.StatusCode)
		}
	}
}

// testApp creates an assets HTTP test app.
func testApp(service port.Service) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	useTestPrincipal(app)
	Register(app, Services{Assets: service, Checker: allowChecker{}})
	return app
}

// allowChecker permits route guards in HTTP adapter tests.
type allowChecker struct{}

// Check returns an allowed decision.
func (allowChecker) Check(context.Context, groupsport.CheckRequest) (groupsport.Decision, error) {
	return groupsport.Decision{Allowed: true, Reason: "test_allowed"}, nil
}

// httptestRequest creates a JSON request.
func httptestRequest(method string, target string, body string) *http.Request {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, _ := http.NewRequest(method, target, reader)
	req.Header.Set(headers.Accept, "application/json")
	req.Header.Set(currentUserIDHeader, uuid.NewString())
	if body != "" {
		req.Header.Set(headers.ContentType, "application/json")
	}
	return req
}

func rawRequest(t *testing.T, method string, target string, body string) *http.Request {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, target, reader)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	return req
}

// testHTTPAsset returns an HTTP test asset.
func testHTTPAsset() domain.Asset {
	return domain.Asset{
		ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Namespace:   "community",
		Path:        "brand",
		Filename:    "logo.png",
		DisplayName: "Logo",
		Visibility:  domain.VisibilityPublic,
		Status:      domain.StatusAvailable,
		StorageKey:  "assets/logo.png",
		Bucket:      "realmkit-assets",
		ContentType: "image/png",
		SizeBytes:   512,
		Version:     1,
	}
}

// httpService is a fake assets service.
type httpService struct {
	asset         domain.Asset
	folders       []string
	namespaces    []string
	err           error
	created       int
	createCommand port.CreateUploadIntentCommand
}

// CreateUploadIntent creates an asset and presigned upload URL.
func (service *httpService) CreateUploadIntent(
	_ context.Context,
	command port.CreateUploadIntentCommand,
) (port.UploadIntent, error) {
	service.created++
	service.createCommand = command
	if service.err != nil {
		return port.UploadIntent{}, service.err
	}
	return port.UploadIntent{Asset: service.asset, Method: "PUT", URL: "https://storage.test/upload", ExpiresAt: time.Now().UTC()}, nil
}

// CompleteUpload confirms the upload object exists.
func (service *httpService) CompleteUpload(context.Context, port.CompleteUploadCommand) (domain.Asset, error) {
	if service.err != nil {
		return domain.Asset{}, service.err
	}
	return service.asset, nil
}

// Get returns one asset.
func (service *httpService) Get(context.Context, uuid.UUID) (domain.Asset, error) {
	if service.err != nil {
		return domain.Asset{}, service.err
	}
	return service.asset, nil
}

// GetURL returns a signed read URL.
func (service *httpService) GetURL(context.Context, uuid.UUID, time.Duration) (string, error) {
	if service.err != nil {
		return "", service.err
	}
	return "https://storage.test/read", nil
}

// List returns matching assets.
func (service *httpService) List(context.Context, port.AssetFilter, pagination.Page) (pagination.Result[domain.Asset], error) {
	if service.err != nil {
		return pagination.Result[domain.Asset]{}, service.err
	}
	return pagination.Result[domain.Asset]{Items: []domain.Asset{service.asset}}, nil
}

// ListNamespaces returns active asset namespaces.
func (service *httpService) ListNamespaces(context.Context) ([]string, error) {
	if service.err != nil {
		return nil, service.err
	}
	return service.namespaces, nil
}

// ListFolders returns direct virtual folder children.
func (service *httpService) ListFolders(context.Context, port.FolderFilter) ([]string, error) {
	if service.err != nil {
		return nil, service.err
	}
	return service.folders, nil
}

// Update changes mutable asset fields.
func (service *httpService) Update(context.Context, port.UpdateAssetCommand) (domain.Asset, error) {
	if service.err != nil {
		return domain.Asset{}, service.err
	}
	return service.asset, nil
}

// Delete soft deletes one asset.
func (service *httpService) Delete(context.Context, port.DeleteAssetCommand) error {
	if service.err != nil && !errors.Is(service.err, port.ErrNotFound) {
		return service.err
	}
	return nil
}
