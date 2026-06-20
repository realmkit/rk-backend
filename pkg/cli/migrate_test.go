package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
	"go.uber.org/zap"
)

// TestMigrateStatusReportsPendingMigration verifies migrate status uses the global runner.
func TestMigrateStatusReportsPendingMigration(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"migrate", "status"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(output.String(), "pending=14") {
		t.Fatalf("output = %q, want pending=14", output.String())
	}
}

// TestMigrateDownRequiresDestructiveConfirmation verifies rollback is guarded.
func TestMigrateDownRequiresDestructiveConfirmation(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"migrate", "down"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("Execute() error = nil, want error")
	}
	if strings.Contains(output.String(), "Usage:") {
		t.Fatalf("output = %q, want no Usage", output.String())
	}
}

// TestMigrateCommandsApplyValidateAndReset verifies stateful migration commands.
func TestMigrateCommandsApplyValidateAndReset(t *testing.T) {
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)
	output, err := executeCommand(t, []string{"migrate", "up"}, deps)
	if err != nil {
		t.Fatalf("up Execute() error = %v", err)
	}
	if !strings.Contains(output, "applied=14 pending=0") {
		t.Fatalf("up output = %q, want applied=14 pending=0", output)
	}
	output, err = executeCommand(t, []string{"migrate", "validate"}, deps)
	if err != nil {
		t.Fatalf("validate Execute() error = %v", err)
	}
	if !strings.Contains(output, "applied=14 pending=0") {
		t.Fatalf("validate output = %q, want applied=14 pending=0", output)
	}
	output, err = executeCommand(t, []string{"migrate", "reset", "--i-understand-this-can-destroy-data"}, deps)
	if err != nil {
		t.Fatalf("reset Execute() error = %v", err)
	}
	if !strings.Contains(output, "applied=0 pending=14") {
		t.Fatalf("reset output = %q, want applied=0 pending=14", output)
	}
}

// TestMigrateRepairRunsWithFlags verifies dirty migration repair wiring.
func TestMigrateRepairRunsWithFlags(t *testing.T) {
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)
	runner := migrations.NewRunner(db, migrations.DefaultSource())
	loaded, err := migrations.Load(migrations.DefaultSource())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := runner.Repair(context.Background(), loaded[0].Version, loaded[0].Checksum, "precondition"); err == nil {
		t.Fatalf("Repair() error = nil, want missing dirty record before setup")
	}
	store := migrations.NewStore(db)
	if err := store.Ensure(context.Background()); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	if err := store.Start(context.Background(), loaded[0], "tester", ""); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	_, err = executeCommand(
		t,
		[]string{"migrate", "repair", "--version", "1", "--checksum", loaded[0].Checksum, "--reason", "manual"},
		deps,
	)
	if err != nil {
		t.Fatalf("repair Execute() error = %v", err)
	}
}
