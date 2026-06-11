package migrations

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// apply executes one migration and records its state.
func (runner Runner) apply(ctx context.Context, migration Migration) error {
	started := time.Now().UTC()
	runner.log.Info(
		"applying database migration",
		zap.Int64("version", migration.Version),
		zap.String("name", migration.Name),
		zap.String("checksum", migration.Checksum),
	)
	if err := runner.store.Start(ctx, migration, runner.executor, runner.appVersion); err != nil {
		return err
	}
	err := runner.execute(ctx, migration)
	if err != nil {
		_ = runner.store.Fail(ctx, migration, started, err)
		return err
	}
	if err := runner.store.Succeed(ctx, migration, started); err != nil {
		return err
	}
	runner.log.Info(
		"database migration applied",
		zap.Int64("version", migration.Version),
		zap.String("name", migration.Name),
		zap.Int64("duration_ms", time.Since(started).Milliseconds()),
	)
	return nil
}

// execute runs migration SQL.
func (runner Runner) execute(ctx context.Context, migration Migration) error {
	sql := migrationSQL(runner.db.Dialector.Name(), migration.UpSQL)
	if !migration.Transaction {
		return runner.db.WithContext(ctx).Exec(sql).Error
	}
	return runner.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Exec(sql).Error
	})
}

// rollback executes a down migration and removes its history row.
func (runner Runner) rollback(ctx context.Context, migration Migration) error {
	sql := migrationSQL(runner.db.Dialector.Name(), migration.DownSQL)
	runner.log.Info(
		"rolling back database migration",
		zap.Int64("version", migration.Version),
		zap.String("name", migration.Name),
	)
	err := runner.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Exec(sql).Error
	})
	if err != nil {
		return err
	}
	return runner.store.Delete(ctx, migration.Version)
}

// migrationSQL returns SQL adapted for the active dialect.
func migrationSQL(dialect string, script string) string {
	if dialect != "sqlite" {
		return script
	}
	script = strings.ReplaceAll(script, "timestamptz", "datetime")
	script = strings.ReplaceAll(script, "uuid", "text")
	script = strings.ReplaceAll(script, "jsonb", "text")
	return stripPostgresOnlyLines(script)
}

// stripPostgresOnlyLines removes clauses unsupported by SQLite test databases.
func stripPostgresOnlyLines(script string) string {
	lines := strings.Split(script, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, "USING gin") || strings.Contains(line, "to_tsvector") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

// validateRecords validates applied records against migration files.
func validateRecords(migrations []Migration, records []Record) error {
	byVersion := map[int64]Migration{}
	for _, migration := range migrations {
		byVersion[migration.Version] = migration
	}
	for _, record := range records {
		if record.Dirty {
			return fmt.Errorf("%w: version %06d", ErrDirty, record.Version)
		}
		migration, ok := byVersion[record.Version]
		if !ok {
			return fmt.Errorf("applied migration %06d has no migration file", record.Version)
		}
		if migration.Checksum != record.Checksum {
			return fmt.Errorf("%w: version %06d", ErrChecksumChanged, record.Version)
		}
	}
	return nil
}

// pendingMigrations returns migrations without applied records.
func pendingMigrations(migrations []Migration, records []Record) []Migration {
	applied := map[int64]struct{}{}
	for _, record := range records {
		if record.Success && !record.Dirty {
			applied[record.Version] = struct{}{}
		}
	}
	var pending []Migration
	for _, migration := range migrations {
		if _, ok := applied[migration.Version]; !ok {
			pending = append(pending, migration)
		}
	}
	return pending
}

// rollbackTargets returns applied migrations in rollback order.
func rollbackTargets(migrations []Migration, records []Record, steps int) []Migration {
	byVersion := map[int64]Migration{}
	for _, migration := range migrations {
		byVersion[migration.Version] = migration
	}
	var targets []Migration
	for index := len(records) - 1; index >= 0 && len(targets) < steps; index-- {
		if migration, ok := byVersion[records[index].Version]; ok {
			targets = append(targets, migration)
		}
	}
	return targets
}

// dirty reports whether any record is dirty.
func dirty(records []Record) bool {
	for _, record := range records {
		if record.Dirty {
			return true
		}
	}
	return false
}

// defaultExecutor returns the default migration executor name.
func defaultExecutor() string {
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		return hostname
	}
	return "gamehub"
}
