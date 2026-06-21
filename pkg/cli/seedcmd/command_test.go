package seedcmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/realmkit/rk-backend/pkg/postgres/seeding"
)

// TestWriteStatusFormatsCountsAndPendingSeeds verifies stable CLI output.
func TestWriteStatusFormatsCountsAndPendingSeeds(t *testing.T) {
	var output bytes.Buffer
	writeStatus(&output, seeding.Status{
		Applied: []seeding.Record{{Version: 1}},
		Pending: []seeding.Seed{{Version: 2, Name: "groups"}},
		Dirty:   true,
	})
	text := output.String()
	for _, want := range []string{"applied=1 pending=1 dirty=true", "pending 000002 groups"} {
		if !strings.Contains(text, want) {
			t.Fatalf("writeStatus() = %q, missing %q", text, want)
		}
	}
}
