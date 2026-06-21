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
	Versions    port.VersionRepository         // Versions stores the versions value.
	Files       port.FileRepository            // Files stores the files value.
	Assets      port.AssetRepository           // Assets stores the assets value.
	Issues      port.ValidationIssueRepository // Issues stores the issues value.
	Signatures  port.SignatureRepository       // Signatures stores the signatures value.
	SigningKeys port.SigningKeyRepository      // SigningKeys stores the signing keys value.
}

// Command requests importing one uploaded theme package.
type Command struct {
	ThemeID          uuid.UUID  // ThemeID stores the theme i d value.
	IdempotencyKey   string     // IdempotencyKey stores the idempotency key value.
	Semver           string     // Semver stores the semver value.
	Label            string     // Label stores the label value.
	ActorUserID      *uuid.UUID // ActorUserID stores the actor user i d value.
	PackageSizeBytes int64      // PackageSizeBytes stores the package size bytes value.
	Package          io.Reader  // Package stores the package value.
}

// Result is the outcome of a package import.
type Result struct {
	Version   domain.ThemeVersion           // Version stores the version value.
	Signature domain.ThemePackageSignature  // Signature stores the signature value.
	Issues    []domain.ThemeValidationIssue // Issues stores the issues value.
	Imported  bool                          // Imported stores the imported value.
	Reused    bool                          // Reused stores the reused value.
}

// Service imports uploaded theme zip packages.
type Service struct {
	repositories Repositories      // repositories stores the repositories value.
	store        storage.Store     // store stores the store value.
	cfg          Config            // cfg stores the cfg value.
	verifier     signatureVerifier // verifier stores the verifier value.
}

// signatureVerifier verifies package signature bytes.
type signatureVerifier interface {
	Verify(context.Context, []byte, []byte) signing.Result
}
