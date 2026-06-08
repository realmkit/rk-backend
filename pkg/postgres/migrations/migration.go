package migrations

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Direction identifies a migration direction.
type Direction string

// Migration directions.
const (
	DirectionUp   Direction = "up"
	DirectionDown Direction = "down"
)

// Source identifies an embedded migration source.
type Source struct {
	// FS is the migration filesystem.
	FS fs.FS

	// Root is the directory containing migration files.
	Root string
}

// Migration contains one globally ordered migration.
type Migration struct {
	// Version is the global migration version.
	Version int64

	// Name is the lower snake case migration name.
	Name string

	// UpPath is the embedded up file path.
	UpPath string

	// DownPath is the embedded down file path.
	DownPath string

	// UpSQL is the forward SQL script.
	UpSQL string

	// DownSQL is the rollback SQL script.
	DownSQL string

	// Checksum is the SHA-256 checksum of UpSQL.
	Checksum string

	// Transaction reports whether UpSQL should run inside a transaction.
	Transaction bool
}

// Load reads, validates, and orders migrations from source.
func Load(source Source) ([]Migration, error) {
	files, err := fs.ReadDir(source.FS, source.Root)
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}

	entries := map[int64]*migrationParts{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		part, err := parseFileName(file.Name())
		if err != nil {
			return nil, err
		}
		content, err := fs.ReadFile(source.FS, path.Join(source.Root, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", file.Name(), err)
		}
		slot := entries[part.version]
		if slot == nil {
			slot = &migrationParts{version: part.version, name: part.name}
			entries[part.version] = slot
		}
		if slot.name != part.name {
			return nil, fmt.Errorf("migration version %06d has conflicting names", part.version)
		}
		slot.add(part, string(content))
	}

	migrations, err := completeMigrations(entries)
	if err != nil {
		return nil, err
	}
	sort.Slice(migrations, func(left int, right int) bool {
		return migrations[left].Version < migrations[right].Version
	})
	return migrations, validateSequence(migrations)
}

// migrationNamePattern matches migration file names.
var migrationNamePattern = regexp.MustCompile(`^([0-9]{6})_([a-z][a-z0-9_]*?)\.(up|down)\.sql$`)

// filePart contains parsed migration file metadata.
type filePart struct {
	version   int64
	name      string
	direction Direction
	fileName  string
}

// migrationParts contains the up and down scripts for one version.
type migrationParts struct {
	version  int64
	name     string
	upPath   string
	downPath string
	upSQL    string
	downSQL  string
}

// parseFileName parses one migration file name.
func parseFileName(fileName string) (filePart, error) {
	matches := migrationNamePattern.FindStringSubmatch(fileName)
	if len(matches) != 4 {
		return filePart{}, fmt.Errorf("invalid migration file name: %s", fileName)
	}
	version, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return filePart{}, fmt.Errorf("parse migration version %s: %w", fileName, err)
	}
	return filePart{version: version, name: matches[2], direction: Direction(matches[3]), fileName: fileName}, nil
}

// add stores one parsed migration file.
func (parts *migrationParts) add(part filePart, content string) {
	switch part.direction {
	case DirectionUp:
		parts.upPath = part.fileName
		parts.upSQL = content
	case DirectionDown:
		parts.downPath = part.fileName
		parts.downSQL = content
	}
}

// completeMigrations validates migration pairs and returns migrations.
func completeMigrations(entries map[int64]*migrationParts) ([]Migration, error) {
	migrations := make([]Migration, 0, len(entries))
	for _, parts := range entries {
		if parts.upSQL == "" {
			return nil, fmt.Errorf("migration %06d missing up script", parts.version)
		}
		if parts.downSQL == "" {
			return nil, fmt.Errorf("migration %06d missing down script", parts.version)
		}
		migrations = append(migrations, Migration{
			Version:     parts.version,
			Name:        parts.name,
			UpPath:      parts.upPath,
			DownPath:    parts.downPath,
			UpSQL:       parts.upSQL,
			DownSQL:     parts.downSQL,
			Checksum:    checksum(parts.upSQL),
			Transaction: transactionEnabled(parts.upSQL),
		})
	}
	return migrations, nil
}

// validateSequence validates migration version ordering.
func validateSequence(migrations []Migration) error {
	for index, migration := range migrations {
		want := int64(index + 1)
		if migration.Version != want {
			return fmt.Errorf("migration sequence gap: got %06d want %06d", migration.Version, want)
		}
	}
	return nil
}

// checksum returns the SHA-256 checksum for sql.
func checksum(sql string) string {
	normalized := strings.ReplaceAll(sql, "\r\n", "\n")
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

// transactionEnabled reports whether script should run in a transaction.
func transactionEnabled(script string) bool {
	for _, line := range strings.Split(script, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		return line != "-- gamehub:transaction false"
	}
	return true
}
