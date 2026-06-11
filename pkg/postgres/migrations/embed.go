package migrations

import "embed"

// Files contains embedded RealmKit SQL migrations.
//
//go:embed migrations/*.sql
var Files embed.FS

// DefaultSource returns the embedded production migration source.
func DefaultSource() Source {
	return Source{FS: Files, Root: "migrations"}
}
