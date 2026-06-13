package seeding

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// apply executes one seed and records its state.
func (runner Runner) apply(ctx context.Context, seed Seed) error {
	started := time.Now().UTC()
	runner.log.Info("applying database seed", zap.Int64("version", seed.Version), zap.String("name", seed.Name))
	if err := runner.store.Start(ctx, seed, runner.executor, runner.appVersion); err != nil {
		return err
	}
	err := runner.execute(ctx, seed)
	if err != nil {
		_ = runner.store.Fail(ctx, seed, started, err)
		return err
	}
	if err := runner.store.Succeed(ctx, seed, started); err != nil {
		return err
	}
	runner.log.Info("database seed applied", zap.Int64("version", seed.Version), zap.String("name", seed.Name))
	return nil
}

// execute runs seed SQL.
func (runner Runner) execute(ctx context.Context, seed Seed) error {
	sql := seedSQL(runner.db.Dialector.Name(), seed.SQL)
	if !seed.Transaction {
		return runner.db.WithContext(ctx).Exec(sql).Error
	}
	return runner.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Exec(sql).Error
	})
}

// grantAdmin creates or restores the seeded administrator membership.
func (runner Runner) grantAdmin(ctx context.Context, userID uuid.UUID) (AdminGrant, error) {
	if err := runner.ensureGrantTarget(ctx, userID); err != nil {
		return AdminGrant{}, err
	}
	created, err := runner.upsertAdminMembership(ctx, userID)
	if err != nil {
		return AdminGrant{}, err
	}
	return AdminGrant{UserID: userID, GroupID: AdminGroupID, Created: created}, nil
}

// ensureGrantTarget verifies the grant can target the seeded admin group and local user.
func (runner Runner) ensureGrantTarget(ctx context.Context, userID uuid.UUID) error {
	userCount, err := countRows(ctx, runner.db, "users", "id = ? AND deleted_at IS NULL", userID)
	if err != nil {
		return err
	}
	if userCount == 0 {
		return fmt.Errorf("%w: %s", ErrUserNotFound, userID)
	}
	groupCount, err := countRows(ctx, runner.db, "groups", "id = ? AND deleted_at IS NULL", AdminGroupID)
	if err != nil {
		return err
	}
	if groupCount == 0 {
		return fmt.Errorf("administrator group seed has not been applied")
	}
	return nil
}

// upsertAdminMembership creates or reactivates the admin group membership.
func (runner Runner) upsertAdminMembership(ctx context.Context, userID uuid.UUID) (bool, error) {
	existing, err := countRows(ctx, runner.db, "group_memberships", "group_id = ? AND user_id = ? AND deleted_at IS NULL", AdminGroupID, userID)
	if err != nil {
		return false, err
	}
	if existing > 0 {
		err = runner.db.WithContext(ctx).Exec(adminMembershipReactivateSQL, AdminGroupID, userID).Error
		return false, err
	}
	membershipID := uuid.NewSHA1(AdminGroupID, []byte(userID.String()))
	err = runner.db.WithContext(ctx).Exec(adminMembershipSQL, membershipID, AdminGroupID, userID).Error
	return existing == 0, err
}

// countRows counts rows matching condition.
func countRows(ctx context.Context, db *gorm.DB, table string, condition string, args ...any) (int64, error) {
	var count int64
	err := db.WithContext(ctx).Table(table).Where(condition, args...).Count(&count).Error
	return count, err
}

// seedSQL returns SQL adapted for the active dialect.
func seedSQL(dialect string, script string) string {
	if dialect != "sqlite" {
		return script
	}
	script = strings.ReplaceAll(script, "timestamptz", "datetime")
	script = strings.ReplaceAll(script, "uuid", "text")
	script = strings.ReplaceAll(script, "jsonb", "text")
	return script
}

// validateRecords validates applied records against seed files.
func validateRecords(seeds []Seed, records []Record) error {
	byVersion := map[int64]Seed{}
	for _, seed := range seeds {
		byVersion[seed.Version] = seed
	}
	for _, record := range records {
		if record.Dirty {
			return fmt.Errorf("%w: version %06d", ErrDirty, record.Version)
		}
		seed, ok := byVersion[record.Version]
		if !ok {
			return fmt.Errorf("applied seed %06d has no seed file", record.Version)
		}
		if seed.Checksum != record.Checksum {
			return fmt.Errorf("%w: version %06d", ErrChecksumChanged, record.Version)
		}
	}
	return nil
}

// pendingSeeds returns seeds without applied records.
func pendingSeeds(seeds []Seed, records []Record) []Seed {
	applied := map[int64]struct{}{}
	for _, record := range records {
		if record.Success && !record.Dirty {
			applied[record.Version] = struct{}{}
		}
	}
	pending := make([]Seed, 0, len(seeds))
	for _, seed := range seeds {
		if _, ok := applied[seed.Version]; !ok {
			pending = append(pending, seed)
		}
	}
	return pending
}

// dirty reports whether any record requires repair.
func dirty(records []Record) bool {
	for _, record := range records {
		if record.Dirty {
			return true
		}
	}
	return false
}

// adminMembershipSQL upserts the seeded administrator membership.
const (
	adminMembershipSQL = `
INSERT INTO group_memberships(
	id, group_id, user_id, status, assigned_by_user_id, assigned_reason,
	starts_at, expires_at, version, created_at, updated_at, deleted_at
) VALUES (?, ?, ?, 'active', NULL, 'Seeded first RealmKit operator grant.', NULL, NULL, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, NULL)
ON CONFLICT(id) DO UPDATE SET
	status = 'active',
	assigned_reason = 'Seeded first RealmKit operator grant.',
	updated_at = CURRENT_TIMESTAMP,
	deleted_at = NULL`

	// adminMembershipReactivateSQL restores an existing active admin grant path.
	adminMembershipReactivateSQL = `
UPDATE group_memberships
SET
	status = 'active',
	assigned_reason = 'Seeded first RealmKit operator grant.',
	updated_at = CURRENT_TIMESTAMP
WHERE group_id = ? AND user_id = ? AND deleted_at IS NULL`
)
