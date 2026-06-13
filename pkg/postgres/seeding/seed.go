package seeding

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

	"github.com/google/uuid"
)

// AdminGroupID is the stable administrator group created by global seeds.
var AdminGroupID = uuid.MustParse("00000000-0000-0000-0000-000000000101")

// Source identifies an embedded seed source.
type Source struct {
	// FS is the seed filesystem.
	FS fs.FS

	// Root is the directory containing seed files.
	Root string
}

// Seed contains one globally ordered data seed.
type Seed struct {
	// Version is the global seed version.
	Version int64

	// Name is the lower snake case seed name.
	Name string

	// Path is the embedded seed file path.
	Path string

	// SQL is the seed SQL script.
	SQL string

	// Checksum is the SHA-256 checksum of SQL.
	Checksum string

	// Transaction reports whether SQL should run inside a transaction.
	Transaction bool
}

// Load reads, validates, and orders seeds from source.
func Load(source Source) ([]Seed, error) {
	files, err := fs.ReadDir(source.FS, source.Root)
	if err != nil {
		return nil, fmt.Errorf("read seeds: %w", err)
	}
	seeds := make([]Seed, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		seed, err := loadSeed(source, file.Name())
		if err != nil {
			return nil, err
		}
		seeds = append(seeds, seed)
	}
	sort.Slice(seeds, func(left int, right int) bool {
		return seeds[left].Version < seeds[right].Version
	})
	return seeds, validateSequence(seeds)
}

// loadSeed loads one seed file from source.
func loadSeed(source Source, fileName string) (Seed, error) {
	version, name, err := parseFileName(fileName)
	if err != nil {
		return Seed{}, err
	}
	content, err := fs.ReadFile(source.FS, path.Join(source.Root, fileName))
	if err != nil {
		return Seed{}, fmt.Errorf("read seed %s: %w", fileName, err)
	}
	sql := string(content)
	return Seed{
		Version:     version,
		Name:        name,
		Path:        fileName,
		SQL:         sql,
		Checksum:    checksum(sql),
		Transaction: transactionEnabled(sql),
	}, nil
}

// seedNamePattern matches seed file names.
var seedNamePattern = regexp.MustCompile(`^([0-9]{6})_([a-z][a-z0-9_]*?)\.up\.sql$`)

// parseFileName parses one seed file name.
func parseFileName(fileName string) (int64, string, error) {
	matches := seedNamePattern.FindStringSubmatch(fileName)
	if len(matches) != 3 {
		return 0, "", fmt.Errorf("invalid seed file name: %s", fileName)
	}
	version, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("parse seed version %s: %w", fileName, err)
	}
	return version, matches[2], nil
}

// validateSequence validates seed version ordering.
func validateSequence(seeds []Seed) error {
	for index, seed := range seeds {
		want := int64(index + 1)
		if seed.Version != want {
			return fmt.Errorf("seed sequence gap: got %06d want %06d", seed.Version, want)
		}
	}
	return nil
}

// checksum returns the SHA-256 checksum for SQL.
func checksum(sql string) string {
	normalized := strings.ReplaceAll(sql, "\r\n", "\n")
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

// transactionEnabled reports whether script should run inside a transaction.
func transactionEnabled(script string) bool {
	for _, line := range strings.Split(script, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		return line != "-- realmkit:transaction false"
	}
	return true
}
