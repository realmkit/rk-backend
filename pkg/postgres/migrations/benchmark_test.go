package migrations

import "testing"

// benchmarkMigrations stores loaded migration benchmark output.
var benchmarkMigrations []Migration

// benchmarkMigrationSQL stores dialect-normalized SQL benchmark output.
var benchmarkMigrationSQL string

// BenchmarkLoadDefaultSource measures embedded migration discovery, pairing, checksum, and sequence validation.
func BenchmarkLoadDefaultSource(b *testing.B) {
	source := DefaultSource()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		migrations, err := Load(source)
		if err != nil {
			b.Fatalf("Load() error = %v", err)
		}
		benchmarkMigrations = migrations
	}
}

// BenchmarkMigrationSQLSQLite measures PostgreSQL-only directive stripping for SQLite test runs.
func BenchmarkMigrationSQLSQLite(b *testing.B) {
	script := `CREATE TABLE realmkit_test(id uuid PRIMARY KEY);
-- postgres-only
CREATE INDEX CONCURRENTLY realmkit_test_id_idx ON realmkit_test(id);
ALTER TABLE realmkit_test ADD COLUMN name text;`

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		benchmarkMigrationSQL = migrationSQL("sqlite", script)
	}
}
