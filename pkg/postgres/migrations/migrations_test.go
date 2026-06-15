package migrations

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestLoadReturnsDefaultMigrations verifies embedded migrations load.
func TestLoadReturnsDefaultMigrations(t *testing.T) {
	migrations, err := Load(DefaultSource())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(migrations) != 9 {
		t.Fatalf("len(migrations) = %d, want 9", len(migrations))
	}
	if migrations[0].Version != 1 || migrations[0].Name != "create_metadata_tables" {
		t.Fatalf("migration[0] = %+v, want metadata version 1", migrations[0])
	}
	if migrations[1].Version != 2 || migrations[1].Name != "create_asset_tables" {
		t.Fatalf("migration[1] = %+v, want assets version 2", migrations[1])
	}
	if migrations[2].Version != 3 || migrations[2].Name != "create_group_permission_tables" {
		t.Fatalf("migration[2] = %+v, want groups version 3", migrations[2])
	}
	if migrations[3].Version != 4 || migrations[3].Name != "create_user_tables" {
		t.Fatalf("migration[3] = %+v, want users version 4", migrations[3])
	}
	if migrations[4].Version != 5 || migrations[4].Name != "create_forum_tables" {
		t.Fatalf("migration[4] = %+v, want forums version 5", migrations[4])
	}
	if migrations[5].Version != 6 || migrations[5].Name != "create_events_and_cronjobs" {
		t.Fatalf("migration[5] = %+v, want events and cronjobs version 6", migrations[5])
	}
	if migrations[6].Version != 7 || migrations[6].Name != "create_punishment_tables" {
		t.Fatalf("migration[6] = %+v, want punishments version 7", migrations[6])
	}
	if migrations[7].Version != 8 || migrations[7].Name != "create_ticket_tables" {
		t.Fatalf("migration[7] = %+v, want tickets version 8", migrations[7])
	}
	if migrations[8].Version != 9 || migrations[8].Name != "create_search_indexes" {
		t.Fatalf("migration[8] = %+v, want search indexes version 9", migrations[8])
	}
}

// TestLoadRejectsVersionGaps verifies the global sequence has no gaps.
func TestLoadRejectsVersionGaps(t *testing.T) {
	_, err := Load(Source{FS: testSource(map[string]string{
		"migrations/000002_skip.up.sql":   "SELECT 1;",
		"migrations/000002_skip.down.sql": "SELECT 1;",
	}), Root: "migrations"})
	if err == nil {
		t.Fatalf("Load() error = nil, want error")
	}
}

// TestLoadRejectsMissingDown verifies every migration has a down file.
func TestLoadRejectsMissingDown(t *testing.T) {
	_, err := Load(Source{FS: testSource(map[string]string{
		"migrations/000001_missing_down.up.sql": "SELECT 1;",
	}), Root: "migrations"})
	if err == nil {
		t.Fatalf("Load() error = nil, want error")
	}
}

// TestRunnerUpAppliesDefaultMigrations verifies default migrations create tables.
func TestRunnerUpAppliesDefaultMigrations(t *testing.T) {
	db := newDB(t)
	status, err := NewRunner(db, DefaultSource()).Up(context.Background())
	if err != nil {
		t.Fatalf("Up() error = %v", err)
	}
	if len(status.Applied) != 9 || len(status.Pending) != 0 {
		t.Fatalf("Status = %+v, want nine applied and no pending", status)
	}
	if !db.Migrator().HasTable("metadata_metafield_definitions") {
		t.Fatalf("metadata_metafield_definitions table missing")
	}
	if !db.Migrator().HasTable("assets") {
		t.Fatalf("assets table missing")
	}
	if !db.Migrator().HasTable("groups") {
		t.Fatalf("groups table missing")
	}
	if !db.Migrator().HasTable("group_memberships") {
		t.Fatalf("group_memberships table missing")
	}
	if !db.Migrator().HasTable("permission_grants") {
		t.Fatalf("permission_grants table missing")
	}
	if !db.Migrator().HasTable("users") {
		t.Fatalf("users table missing")
	}
	if !db.Migrator().HasTable("forums") {
		t.Fatalf("forums table missing")
	}
	if !db.Migrator().HasTable("forum_categories") {
		t.Fatalf("forum_categories table missing")
	}
	if !db.Migrator().HasTable("forum_stats") {
		t.Fatalf("forum_stats table missing")
	}
	if !db.Migrator().HasTable("event_outbox") {
		t.Fatalf("event_outbox table missing")
	}
	if !db.Migrator().HasTable("cronjob_definitions") {
		t.Fatalf("cronjob_definitions table missing")
	}
	if !db.Migrator().HasTable("punishments") {
		t.Fatalf("punishments table missing")
	}
	if !db.Migrator().HasTable("tickets") {
		t.Fatalf("tickets table missing")
	}
	if !db.Migrator().HasTable("ticket_definitions") {
		t.Fatalf("ticket_definitions table missing")
	}
}

// TestRunnerUpIsIdempotent verifies re-running migrations has no pending work.
func TestRunnerUpIsIdempotent(t *testing.T) {
	db := newDB(t)
	runner := NewRunner(db, DefaultSource())
	if _, err := runner.Up(context.Background()); err != nil {
		t.Fatalf("first Up() error = %v", err)
	}
	status, err := runner.Up(context.Background())
	if err != nil {
		t.Fatalf("second Up() error = %v", err)
	}
	if len(status.Pending) != 0 {
		t.Fatalf("Pending = %d, want 0", len(status.Pending))
	}
}

// TestRunnerValidateRejectsChecksumChanges verifies applied migrations are immutable.
func TestRunnerValidateRejectsChecksumChanges(t *testing.T) {
	db := newDB(t)
	source := Source{FS: testSource(map[string]string{
		"migrations/000001_first.up.sql":   "CREATE TABLE first_table(id text PRIMARY KEY);",
		"migrations/000001_first.down.sql": "DROP TABLE first_table;",
	}), Root: "migrations"}
	if _, err := NewRunner(db, source).Up(context.Background()); err != nil {
		t.Fatalf("Up() error = %v", err)
	}
	changed := Source{FS: testSource(map[string]string{
		"migrations/000001_first.up.sql":   "CREATE TABLE first_table(id text PRIMARY KEY, name text);",
		"migrations/000001_first.down.sql": "DROP TABLE first_table;",
	}), Root: "migrations"}

	_, err := NewRunner(db, changed).Validate(context.Background())
	if !errors.Is(err, ErrChecksumChanged) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrChecksumChanged)
	}
}

// TestRunnerRepairClearsDirtyState verifies explicit repair clears dirty history.
func TestRunnerRepairClearsDirtyState(t *testing.T) {
	db := newDB(t)
	runner := NewRunner(db, DefaultSource())
	if err := runner.store.Ensure(context.Background()); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	migrations, err := Load(DefaultSource())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := runner.store.Start(context.Background(), migrations[0], "test", ""); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if _, err := runner.Validate(context.Background()); !errors.Is(err, ErrDirty) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrDirty)
	}
	if err := runner.Repair(context.Background(), migrations[0].Version, migrations[0].Checksum, "manual test repair"); err != nil {
		t.Fatalf("Repair() error = %v", err)
	}
	if _, err := runner.Validate(context.Background()); err != nil {
		t.Fatalf("Validate() after repair error = %v", err)
	}
}

// TestRunnerDownRollsBackMigration verifies down migration removes applied schema.
func TestRunnerDownRollsBackMigration(t *testing.T) {
	db := newDB(t)
	runner := NewRunner(db, DefaultSource())
	if _, err := runner.Up(context.Background()); err != nil {
		t.Fatalf("Up() error = %v", err)
	}

	status, err := runner.Down(context.Background(), 9)
	if err != nil {
		t.Fatalf("Down() error = %v", err)
	}
	if len(status.Applied) != 0 || len(status.Pending) != 9 {
		t.Fatalf("Status = %+v, want no applied and nine pending", status)
	}
	if db.Migrator().HasTable("metadata_metafield_definitions") {
		t.Fatalf("metadata_metafield_definitions table exists after Down()")
	}
	if db.Migrator().HasTable("assets") {
		t.Fatalf("assets table exists after Down()")
	}
	if db.Migrator().HasTable("groups") {
		t.Fatalf("groups table exists after Down()")
	}
	if db.Migrator().HasTable("users") {
		t.Fatalf("users table exists after Down()")
	}
	if db.Migrator().HasTable("forums") {
		t.Fatalf("forums table exists after Down()")
	}
	if db.Migrator().HasTable("punishments") {
		t.Fatalf("punishments table exists after Down()")
	}
}

// TestRunnerResetEmptyDatabaseIsNoop verifies reset is safe on empty state.
func TestRunnerResetEmptyDatabaseIsNoop(t *testing.T) {
	status, err := NewRunner(newDB(t), DefaultSource()).Reset(context.Background())
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}
	if len(status.Applied) != 0 {
		t.Fatalf("Applied = %d, want 0", len(status.Applied))
	}
}

// TestRunnerOptionsConfigureMetadata verifies runner options are retained.
func TestRunnerOptionsConfigureMetadata(t *testing.T) {
	log := zap.NewNop()
	runner := NewRunner(newDB(t), DefaultSource(), WithLogger(log), WithExecutor("tester"), WithAppVersion("v1.2.3"))
	if runner.log != log {
		t.Fatalf("log = %p, want %p", runner.log, log)
	}
	if runner.executor != "tester" {
		t.Fatalf("executor = %q, want tester", runner.executor)
	}
	if runner.appVersion != "v1.2.3" {
		t.Fatalf("appVersion = %q, want v1.2.3", runner.appVersion)
	}
}

// TestStoreFailMarksDirtyRecord verifies failed migrations remain dirty.
func TestStoreFailMarksDirtyRecord(t *testing.T) {
	db := newDB(t)
	store := NewStore(db)
	if err := store.Ensure(context.Background()); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	migrations, err := Load(DefaultSource())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := store.Start(context.Background(), migrations[0], "tester", "v1"); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	started := time.Now().UTC().Add(-time.Second)
	if err := store.Fail(context.Background(), migrations[0], started, errors.New("boom")); err != nil {
		t.Fatalf("Fail() error = %v", err)
	}

	records, err := store.Applied(context.Background())
	if err != nil {
		t.Fatalf("Applied() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}
	if records[0].Success || !records[0].Dirty || records[0].Error != "boom" {
		t.Fatalf("record = %+v, want failed dirty boom record", records[0])
	}
}

// newDB creates a migrated test database.
func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	return db
}

// testSource creates a migration source filesystem.
func testSource(files map[string]string) fs.FS {
	source := fstest.MapFS{}
	for name, content := range files {
		source[name] = &fstest.MapFile{Data: []byte(content)}
	}
	return source
}
