// Package validation owns theme syntax, dependency, security, and coverage checks.
package validation

import (
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// Repositories contains persistence ports required by validation.
type Repositories struct {
	Versions port.VersionRepository
	Files    port.FileRepository
	Issues   port.ValidationIssueRepository
}

// Command requests static validation for one version.
type Command struct {
	VersionID   uuid.UUID
	ActorUserID *uuid.UUID
}

// Result contains a static validation report.
type Result struct {
	Version      domain.ThemeVersion
	Issues       []domain.ThemeValidationIssue
	ManifestJSON []byte
}

// Service validates theme files and writes version diagnostics.
type Service struct {
	repositories Repositories
}

// coverageEntry describes one route template requirement.
type coverageEntry struct {
	Route   domain.RouteKind `json:"route"`
	Path    domain.FilePath  `json:"path"`
	Present bool             `json:"present"`
}

// dependencyReport describes extracted file dependencies.
type dependencyReport struct {
	Sections []string `json:"sections"`
	Snippets []string `json:"snippets"`
	Assets   []string `json:"assets"`
}
