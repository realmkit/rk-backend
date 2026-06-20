package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrPublishedVersionImmutable reports a write to a published version.
var ErrPublishedVersionImmutable = errors.New("published theme version is immutable")

// Key identifies operator-facing theme records.
type Key string

// FilePath is a normalized slash-separated theme package path.
type FilePath string

// Digest is a lowercase hex SHA-256 digest.
type Digest string

// Theme is a versioned theme family.
type Theme struct {
	ID          uuid.UUID
	Key         Key
	Name        string
	Description string
	Status      ThemeStatus
	Version     uint64
	CreatedBy   *uuid.UUID
	UpdatedBy   *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ThemeVersion is one immutable package revision once published.
type ThemeVersion struct {
	ID                 uuid.UUID
	ThemeID            uuid.UUID
	Semver             string
	Label              string
	Status             VersionStatus
	SourceKind         SourceKind
	SourceReference    string
	PackageStorageKey  string
	PackageSizeBytes   int64
	ManifestJSON       []byte
	SettingsSchemaJSON []byte
	SettingsDataJSON   []byte
	IntegritySHA256    Digest
	PublishedAt        *time.Time
	PublishedBy        *uuid.UUID
	Version            uint64
	CreatedBy          *uuid.UUID
	UpdatedBy          *uuid.UUID
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// EnsureEditable returns an error when the version cannot be mutated.
func (version ThemeVersion) EnsureEditable() error {
	if version.Status == VersionStatusPublished || version.PublishedAt != nil {
		return ErrPublishedVersionImmutable
	}
	return nil
}

// ThemeFile stores one logical file in a theme version.
type ThemeFile struct {
	ID             uuid.UUID
	VersionID      uuid.UUID
	Kind           FileKind
	Path           FilePath
	ContentSHA256  Digest
	ContentStorage string
	ContentText    string
	SizeBytes      int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ThemeAsset stores derived public serving metadata for a theme asset file.
type ThemeAsset struct {
	ID             uuid.UUID
	VersionID      uuid.UUID
	FileID         uuid.UUID
	Path           FilePath
	ContentType    string
	SizeBytes      int64
	ContentSHA256  Digest
	StorageKey     string
	PublicURL      string
	IntegrityValue string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ThemeActivation points one environment at an immutable theme version.
type ThemeActivation struct {
	ID          uuid.UUID
	ThemeID     uuid.UUID
	VersionID   uuid.UUID
	Environment ActivationEnvironment
	IsCurrent   bool
	Reason      string
	ActivatedBy *uuid.UUID
	ActivatedAt time.Time
	CreatedAt   time.Time
}

// ThemeValidationIssue stores a structured validation diagnostic.
type ThemeValidationIssue struct {
	ID        uuid.UUID
	VersionID uuid.UUID
	Severity  ValidationSeverity
	Code      ValidationIssueCode
	Path      FilePath
	Message   string
	Line      int
	Column    int
	Details   []byte
	CreatedAt time.Time
}

// ThemePackageSignature stores detached package signature verification data.
type ThemePackageSignature struct {
	ID                 uuid.UUID
	VersionID          uuid.UUID
	KeyID              string
	Algorithm          SignatureAlgorithm
	VerificationStatus SignatureVerificationStatus
	Signature          string
	SignedManifestHash Digest
	VerifiedAt         *time.Time
	CreatedAt          time.Time
}

// ThemeSigningKey is a trusted package signing public key.
type ThemeSigningKey struct {
	ID          uuid.UUID
	KeyID       string
	Algorithm   SignatureAlgorithm
	PublicKey   string
	TrustLevel  SigningKeyTrustLevel
	Status      SigningKeyStatus
	Source      SigningKeySource
	NotBefore   *time.Time
	NotAfter    *time.Time
	CreatedBy   *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	RetiredAt   *time.Time
	RevokedAt   *time.Time
	Description string
}

// ThemePreviewToken grants short-lived preview access to one version.
type ThemePreviewToken struct {
	ID            uuid.UUID
	VersionID     uuid.UUID
	TokenHash     string
	PersonaKind   PreviewPersonaKind
	PersonaSource PreviewPersonaSource
	PersonaUserID *uuid.UUID
	ExpiresAt     time.Time
	CreatedBy     *uuid.UUID
	CreatedAt     time.Time
	RevokedAt     *time.Time
}
