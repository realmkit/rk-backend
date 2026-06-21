// Package validation owns theme syntax, dependency, security, and coverage checks.
package validation

import (
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// Repositories contains persistence ports required by validation.
type Repositories struct {
	Versions port.VersionRepository         // Versions stores the versions value.
	Files    port.FileRepository            // Files stores the files value.
	Issues   port.ValidationIssueRepository // Issues stores the issues value.
}

// Command requests static validation for one version.
type Command struct {
	VersionID   uuid.UUID  // VersionID stores the version i d value.
	ActorUserID *uuid.UUID // ActorUserID stores the actor user i d value.
}

// Result contains a static validation report.
type Result struct {
	Version      domain.ThemeVersion           // Version stores the version value.
	Issues       []domain.ThemeValidationIssue // Issues stores the issues value.
	ManifestJSON []byte                        // ManifestJSON stores the manifest j s o n value.
}

// Service validates theme files and writes version diagnostics.
type Service struct {
	repositories Repositories // repositories stores the repositories value.
}

// coverageEntry describes one route template requirement.
type coverageEntry struct {
	Route   domain.RouteKind `json:"route"`   // Route stores the route value.
	Path    domain.FilePath  `json:"path"`    // Path stores the path value.
	Present bool             `json:"present"` // Present stores the present value.
}

// dependencyReport describes extracted file dependencies.
type dependencyReport struct {
	Sections []string `json:"sections"` // Sections stores the sections value.
	Snippets []string `json:"snippets"` // Snippets stores the snippets value.
	Assets   []string `json:"assets"`   // Assets stores the assets value.
}
