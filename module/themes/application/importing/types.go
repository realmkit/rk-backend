package importing

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/application/signing"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
	"github.com/realmkit/rk-backend/pkg/storage"
)

// Repositories contains persistence ports required by package import.
type Repositories struct {
	Versions    port.VersionRepository
	Files       port.FileRepository
	Assets      port.AssetRepository
	Issues      port.ValidationIssueRepository
	Signatures  port.SignatureRepository
	SigningKeys port.SigningKeyRepository
}

// Command requests importing one uploaded theme package.
type Command struct {
	ThemeID          uuid.UUID
	IdempotencyKey   string
	Semver           string
	Label            string
	ActorUserID      *uuid.UUID
	PackageSizeBytes int64
	Package          io.Reader
}

// Result is the outcome of a package import.
type Result struct {
	Version   domain.ThemeVersion
	Signature domain.ThemePackageSignature
	Issues    []domain.ThemeValidationIssue
	Imported  bool
	Reused    bool
}

// Service imports uploaded theme zip packages.
type Service struct {
	repositories Repositories
	store        storage.Store
	cfg          Config
	verifier     signatureVerifier
}

// signatureVerifier verifies package signature bytes.
type signatureVerifier interface {
	Verify(context.Context, []byte, []byte) signing.Result
}
