package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// TestRootCommandShowsHelpByDefault verifies no-arg execution shows commands.
func TestRootCommandShowsHelpByDefault(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(nil)

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(output.String(), "start") || !strings.Contains(output.String(), "seed") ||
		!strings.Contains(output.String(), "forums") {
		t.Fatalf("output = %q, want start, seed, and forums commands", output.String())
	}
}

// TestRootCommandHelpPrintsUsage verifies help is the usage-printing path.
func TestRootCommandHelpPrintsUsage(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"help"})

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(output.String(), "Usage:") {
		t.Fatalf("output = %q, want Usage", output.String())
	}
}

// TestRunPrintsHelp verifies the production entry point exposes command help.
func TestRunPrintsHelp(t *testing.T) {
	activeLogger := zap.NewNop()
	if err := Run(context.Background(), []string{"help"}, &activeLogger); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

// TestRootCommandErrorDoesNotPrintUsage verifies errors do not include usage output.
func TestRootCommandErrorDoesNotPrintUsage(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"migrate", "repair"})

	if err := cmd.ExecuteContext(context.Background()); err == nil {
		t.Fatalf("Execute() error = nil, want error")
	}
	if strings.Contains(output.String(), "Usage:") {
		t.Fatalf("output = %q, want no Usage", output.String())
	}
}
