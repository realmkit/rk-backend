package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	realmkitredis "github.com/realmkit/rk-backend/pkg/redis"
	goredis "github.com/redis/go-redis/v9"
)

// TestSeedCommandsApplyAndValidate verifies stateful data seed commands.
func TestSeedCommandsApplyAndValidate(t *testing.T) {
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)
	if _, err := executeCommand(t, []string{"migrate", "up"}, deps); err != nil {
		t.Fatalf("migrate up error = %v", err)
	}
	output, err := executeCommand(t, []string{"seed", "up"}, deps)
	if err != nil {
		t.Fatalf("seed up error = %v", err)
	}
	if !strings.Contains(output, "applied=4 pending=0") {
		t.Fatalf("seed up output = %q, want applied=4 pending=0", output)
	}
	output, err = executeCommand(t, []string{"seed", "validate"}, deps)
	if err != nil {
		t.Fatalf("seed validate error = %v", err)
	}
	if !strings.Contains(output, "applied=4 pending=0") {
		t.Fatalf("seed validate output = %q, want applied=4 pending=0", output)
	}
}

// TestSeedDryRunReportsPendingSeeds verifies dry-run does not apply seeds.
func TestSeedDryRunReportsPendingSeeds(t *testing.T) {
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)
	if _, err := executeCommand(t, []string{"migrate", "up"}, deps); err != nil {
		t.Fatalf("migrate up error = %v", err)
	}
	output, err := executeCommand(t, []string{"seed", "dry-run"}, deps)
	if err != nil {
		t.Fatalf("seed dry-run error = %v", err)
	}
	if !strings.Contains(output, "applied=0 pending=4") {
		t.Fatalf("dry-run output = %q, want pending seeds", output)
	}
}

// TestSeedGrantAdminCommandAssignsMembership verifies first-operator grants.
func TestSeedGrantAdminCommandAssignsMembership(t *testing.T) {
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)
	userID := "00000000-0000-0000-0000-000000000901"
	if _, err := executeCommand(t, []string{"migrate", "up"}, deps); err != nil {
		t.Fatalf("migrate up error = %v", err)
	}
	if _, err := executeCommand(t, []string{"seed", "up"}, deps); err != nil {
		t.Fatalf("seed up error = %v", err)
	}
	insertUser := `
	INSERT INTO users(id, status, avatar_asset_id, first_seen_at, last_seen_at, version, created_at, updated_at, deleted_at)
	VALUES(?, 'active', NULL, CURRENT_TIMESTAMP, NULL, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, NULL)`
	if err := db.Exec(insertUser, userID).Error; err != nil {
		t.Fatalf("insert user error = %v", err)
	}
	output, err := executeCommand(t, []string{"seed", "grant-admin", "--user-id", userID}, deps)
	if err != nil {
		t.Fatalf("grant-admin error = %v", err)
	}
	if !strings.Contains(output, "created=true") {
		t.Fatalf("grant output = %q, want created=true", output)
	}
}

// TestForumStatsVerifyCommandReportsCleanCounters verifies operational stats command wiring.
func TestForumStatsVerifyCommandReportsCleanCounters(t *testing.T) {
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)
	if _, err := executeCommand(t, []string{"migrate", "up"}, deps); err != nil {
		t.Fatalf("migrate up error = %v", err)
	}
	output, err := executeCommand(t, []string{"forums", "stats", "verify"}, deps)
	if err != nil {
		t.Fatalf("forums stats verify error = %v", err)
	}
	if !strings.Contains(output, "mismatches=0 repaired=false") {
		t.Fatalf("output = %q, want clean stats report", output)
	}
}

// TestForumCacheClearCommandUsesRedis verifies Redis-backed cache clearing command wiring.
func TestForumCacheClearCommandUsesRedis(t *testing.T) {
	server := miniredis.RunT(t)
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)
	deps.openRedis = func(context.Context, realmkitredis.Config) (*goredis.Client, error) {
		return goredis.NewClient(&goredis.Options{Addr: server.Addr()}), nil
	}
	if _, err := executeCommand(t, []string{"migrate", "up"}, deps); err != nil {
		t.Fatalf("migrate up error = %v", err)
	}
	output, err := executeCommand(t, []string{"forums", "cache", "clear"}, deps)
	if err != nil {
		t.Fatalf("forums cache clear error = %v", err)
	}
	if !strings.Contains(output, "forum caches cleared") {
		t.Fatalf("output = %q, want cache clear message", output)
	}
}
