package harness

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/realmkit/rk-backend/pkg/api/headers"
)

// JSONRequest creates a JSON request with the service's common headers.
func JSONRequest(method string, target string, body string) *http.Request {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, target, reader)
	request.Header.Set(headers.Accept, "application/json")
	if body != "" {
		request.Header.Set(headers.ContentType, "application/json")
	}
	return request
}

// ResponseBody reads response content for assertion messages.
func ResponseBody(t *testing.T, response *http.Response) string {
	t.Helper()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return "unreadable body: " + err.Error()
	}
	return string(payload)
}

// Test executes an HTTP request against the in-process server.
func (ecosystem *Ecosystem) Test(t *testing.T, request *http.Request) *http.Response {
	t.Helper()

	response, err := ecosystem.App.Test(request, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	t.Cleanup(func() {
		if err := response.Body.Close(); err != nil {
			t.Fatalf("Body.Close() error = %v", err)
		}
	})
	return response
}
