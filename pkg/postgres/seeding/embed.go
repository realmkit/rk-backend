package seeding

import "embed"

// Files contains embedded RealmKit SQL seeds.
//
//go:embed seeds/*.sql
var Files embed.FS

// DefaultSource returns the embedded production seed source.
func DefaultSource() Source {
	return Source{FS: Files, Root: "seeds"}
}
