package seeding

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestLoadValidatesSequence verifies embedded seed ordering and checksums.
func TestLoadValidatesSequence(t *testing.T) {
	seeds, err := Load(DefaultSource())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(seeds) != 3 {
		t.Fatalf("seeds = %d, want 3", len(seeds))
	}
	if seeds[0].Version != 1 || seeds[0].Checksum == "" {
		t.Fatalf("first seed = %+v, want version and checksum", seeds[0])
	}
}

// TestLoadRejectsSequenceGaps verifies data seeds must be globally ordered.
func TestLoadRejectsSequenceGaps(t *testing.T) {
	source := Source{
		FS: fstest.MapFS{
			"seeds/000002_late.up.sql": {Data: []byte("SELECT 1;")},
		},
		Root: "seeds",
	}

	_, err := Load(source)
	if err == nil || !strings.Contains(err.Error(), "sequence gap") {
		t.Fatalf("Load() error = %v, want sequence gap", err)
	}
}

// TestRunnerAppliesSeedsIdempotently verifies seeds can be replayed safely.
func TestRunnerAppliesSeedsIdempotently(t *testing.T) {
	db := migratedDB(t)
	runner := NewRunner(db, DefaultSource())

	status, err := runner.Up(context.Background())
	if err != nil {
		t.Fatalf("Up() error = %v", err)
	}
	if len(status.Applied) != 3 || len(status.Pending) != 0 {
		t.Fatalf("status = %+v, want all applied", status)
	}
	status, err = runner.Up(context.Background())
	if err != nil {
		t.Fatalf("second Up() error = %v", err)
	}
	if len(status.Applied) != 3 || len(status.Pending) != 0 {
		t.Fatalf("second status = %+v, want no pending", status)
	}
	assertCount(t, db, "groups", "key = ?", "administrator")
	assertCount(t, db, "permission_grants", "action = ?", "groups.manage_permissions")
	assertCount(t, db, "forum_permission_grants", "id = ?", "00000000-0000-0000-0000-000000000401")
	assertCount(t, db, "forums", "key = ?", "announcements")
}

// TestRunnerValidateDetectsDirtySeed verifies failed seed state blocks validation.
func TestRunnerValidateDetectsDirtySeed(t *testing.T) {
	db := migratedDB(t)
	runner := NewRunner(db, DefaultSource())
	seeds, err := Load(DefaultSource())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := runner.store.Ensure(context.Background()); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	if err := runner.store.Start(context.Background(), seeds[0], "tester", ""); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	_, err = runner.Validate(context.Background())
	if !errors.Is(err, ErrDirty) {
		t.Fatalf("Validate() error = %v, want ErrDirty", err)
	}
}

// TestGrantAdminRequiresLocalUser verifies admin grants target known users.
func TestGrantAdminRequiresLocalUser(t *testing.T) {
	db := migratedDB(t)
	runner := NewRunner(db, DefaultSource())
	if _, err := runner.Up(context.Background()); err != nil {
		t.Fatalf("Up() error = %v", err)
	}

	_, err := runner.GrantAdmin(context.Background(), uuid.New())
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("GrantAdmin() error = %v, want ErrUserNotFound", err)
	}
}

// TestGrantAdminCreatesMembership verifies first-operator grant persistence.
func TestGrantAdminCreatesMembership(t *testing.T) {
	db := migratedDB(t)
	runner := NewRunner(db, DefaultSource())
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000901")
	if _, err := runner.Up(context.Background()); err != nil {
		t.Fatalf("Up() error = %v", err)
	}
	if err := insertUser(db, userID); err != nil {
		t.Fatalf("insertUser() error = %v", err)
	}

	grant, err := runner.GrantAdmin(context.Background(), userID)
	if err != nil {
		t.Fatalf("GrantAdmin() error = %v", err)
	}
	if !grant.Created || grant.GroupID != AdminGroupID {
		t.Fatalf("grant = %+v, want created admin grant", grant)
	}
	grant, err = runner.GrantAdmin(context.Background(), userID)
	if err != nil {
		t.Fatalf("second GrantAdmin() error = %v", err)
	}
	if grant.Created {
		t.Fatalf("second grant = %+v, want existing grant", grant)
	}
	assertCount(t, db, "group_memberships", "group_id = ? AND user_id = ? AND deleted_at IS NULL", AdminGroupID, userID)
}

// migratedDB opens SQLite and applies global migrations.
func migratedDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("migrate Up() error = %v", err)
	}
	return db
}

// insertUser creates a minimal local user row.
func insertUser(db *gorm.DB, userID uuid.UUID) error {
	insert := `
INSERT INTO users(id, status, avatar_asset_id, first_seen_at, last_seen_at, version, created_at, updated_at, deleted_at)
VALUES(?, 'active', NULL, CURRENT_TIMESTAMP, NULL, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, NULL)`
	return db.Exec(insert, userID).Error
}

// assertCount verifies a table has one matching row.
func assertCount(t *testing.T, db *gorm.DB, table string, condition string, args ...any) {
	t.Helper()
	var count int64
	if err := db.Table(table).Where(condition, args...).Count(&count).Error; err != nil {
		t.Fatalf("count %s error = %v", table, err)
	}
	if count != 1 {
		t.Fatalf("%s count = %d, want 1", table, count)
	}
}
