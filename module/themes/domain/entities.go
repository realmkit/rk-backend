package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrPublishedVersionImmutable reports writing to a published version.
var ErrPublishedVersionImmutable = errors.New("published theme version is immutable")

// ErrVersionNotPublishable reports a version that cannot be activated.
var ErrVersionNotPublishable = errors.New("theme version is not publishable")

// ErrInvalidActivation reports an incomplete or unsupported activation.
var ErrInvalidActivation = errors.New("theme activation is invalid")

// ErrSigningKeyInactive reports a key that cannot verify packages.
var ErrSigningKeyInactive = errors.New("theme signing key is inactive")

// ErrPreviewTokenInvalid reports an unusable preview token.
var ErrPreviewTokenInvalid = errors.New("theme preview token is invalid")

// ErrPreviewTokenExpired reports an expired preview token.
var ErrPreviewTokenExpired = errors.New("theme preview token expired")

// ErrPreviewTokenRevoked reports a revoked preview token.
var ErrPreviewTokenRevoked = errors.New("theme preview token revoked")

// Key identifies operator-facing theme records.
type Key string

// FilePath is a normalized slash-separated theme package path.
type FilePath string

// Digest is a lowercase hex SHA-256 digest.
type Digest string

// Theme is a versioned theme family.
type Theme struct {
	ID          uuid.UUID   // ID stores the i d value.
	Key         Key         // Key stores the key value.
	Name        string      // Name stores the name value.
	Description string      // Description stores the description value.
	Status      ThemeStatus // Status stores the status value.
	Version     uint64      // Version stores the version value.
	CreatedBy   *uuid.UUID  // CreatedBy stores the created by value.
	UpdatedBy   *uuid.UUID  // UpdatedBy stores the updated by value.
	CreatedAt   time.Time   // CreatedAt stores the created at value.
	UpdatedAt   time.Time   // UpdatedAt stores the updated at value.
}

// ThemeVersion is one immutable package revision once published.
type ThemeVersion struct {
	ID                 uuid.UUID     // ID stores the i d value.
	ThemeID            uuid.UUID     // ThemeID stores the theme i d value.
	Semver             string        // Semver stores the semver value.
	Label              string        // Label stores the label value.
	Status             VersionStatus // Status stores the status value.
	SourceKind         SourceKind    // SourceKind stores the source kind value.
	SourceReference    string        // SourceReference stores the source reference value.
	PackageStorageKey  string        // PackageStorageKey stores the package storage key value.
	PackageSizeBytes   int64         // PackageSizeBytes stores the package size bytes value.
	ManifestJSON       []byte        // ManifestJSON stores the manifest j s o n value.
	SettingsSchemaJSON []byte        // SettingsSchemaJSON stores the settings schema j s o n value.
	SettingsDataJSON   []byte        // SettingsDataJSON stores the settings data j s o n value.
	IntegritySHA256    Digest        // IntegritySHA256 stores the integrity s h a256 value.
	PublishedAt        *time.Time    // PublishedAt stores the published at value.
	PublishedBy        *uuid.UUID    // PublishedBy stores the published by value.
	Version            uint64        // Version stores the version value.
	CreatedBy          *uuid.UUID    // CreatedBy stores the created by value.
	UpdatedBy          *uuid.UUID    // UpdatedBy stores the updated by value.
	CreatedAt          time.Time     // CreatedAt stores the created at value.
	UpdatedAt          time.Time     // UpdatedAt stores the updated at value.
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
	ID             uuid.UUID // ID stores the i d value.
	VersionID      uuid.UUID // VersionID stores the version i d value.
	Kind           FileKind  // Kind stores the kind value.
	Path           FilePath  // Path stores the path value.
	ContentSHA256  Digest    // ContentSHA256 stores the content s h a256 value.
	ContentStorage string    // ContentStorage stores the content storage value.
	ContentText    string    // ContentText stores the content text value.
	SizeBytes      int64     // SizeBytes stores the size bytes value.
	CreatedAt      time.Time // CreatedAt stores the created at value.
	UpdatedAt      time.Time // UpdatedAt stores the updated at value.
}

// ThemeAsset stores derived public serving metadata for a theme asset file.
type ThemeAsset struct {
	ID             uuid.UUID // ID stores the i d value.
	VersionID      uuid.UUID // VersionID stores the version i d value.
	FileID         uuid.UUID // FileID stores the file i d value.
	Path           FilePath  // Path stores the path value.
	ContentType    string    // ContentType stores the content type value.
	SizeBytes      int64     // SizeBytes stores the size bytes value.
	ContentSHA256  Digest    // ContentSHA256 stores the content s h a256 value.
	StorageKey     string    // StorageKey stores the storage key value.
	PublicURL      string    // PublicURL stores the public u r l value.
	IntegrityValue string    // IntegrityValue stores the integrity value value.
	CreatedAt      time.Time // CreatedAt stores the created at value.
	UpdatedAt      time.Time // UpdatedAt stores the updated at value.
}

// ThemeActivation points one environment at an immutable theme version.
type ThemeActivation struct {
	ID               uuid.UUID             // ID stores the i d value.
	ThemeID          uuid.UUID             // ThemeID stores the theme i d value.
	VersionID        uuid.UUID             // VersionID stores the version i d value.
	Environment      ActivationEnvironment // Environment stores the environment value.
	IsCurrent        bool                  // IsCurrent stores the is current value.
	Reason           string                // Reason stores the reason value.
	SettingsDataJSON []byte                // SettingsDataJSON stores the settings data j s o n value.
	ActivatedBy      *uuid.UUID            // ActivatedBy stores the activated by value.
	ActivatedAt      time.Time             // ActivatedAt stores the activated at value.
	CreatedAt        time.Time             // CreatedAt stores the created at value.
}

// ThemeValidationIssue stores a structured validation diagnostic.
type ThemeValidationIssue struct {
	ID        uuid.UUID           // ID stores the i d value.
	VersionID uuid.UUID           // VersionID stores the version i d value.
	Severity  ValidationSeverity  // Severity stores the severity value.
	Code      ValidationIssueCode // Code stores the code value.
	Path      FilePath            // Path stores the path value.
	Message   string              // Message stores the message value.
	Line      int                 // Line stores the line value.
	Column    int                 // Column stores the column value.
	Details   []byte              // Details stores the details value.
	CreatedAt time.Time           // CreatedAt stores the created at value.
}

// ThemePackageSignature stores detached package signature verification data.
type ThemePackageSignature struct {
	ID                 uuid.UUID                   // ID stores the i d value.
	VersionID          uuid.UUID                   // VersionID stores the version i d value.
	KeyID              string                      // KeyID stores the key i d value.
	Algorithm          SignatureAlgorithm          // Algorithm stores the algorithm value.
	VerificationStatus SignatureVerificationStatus // VerificationStatus stores the verification status value.
	Signature          string                      // Signature stores the signature value.
	SignedManifestHash Digest                      // SignedManifestHash stores the signed manifest hash value.
	VerifiedAt         *time.Time                  // VerifiedAt stores the verified at value.
	CreatedAt          time.Time                   // CreatedAt stores the created at value.
}

// ThemeSigningKey is a trusted package signing public key.
type ThemeSigningKey struct {
	ID          uuid.UUID            // ID stores the i d value.
	KeyID       string               // KeyID stores the key i d value.
	Algorithm   SignatureAlgorithm   // Algorithm stores the algorithm value.
	PublicKey   string               // PublicKey stores the public key value.
	TrustLevel  SigningKeyTrustLevel // TrustLevel stores the trust level value.
	Status      SigningKeyStatus     // Status stores the status value.
	Source      SigningKeySource     // Source stores the source value.
	NotBefore   *time.Time           // NotBefore stores the not before value.
	NotAfter    *time.Time           // NotAfter stores the not after value.
	CreatedBy   *uuid.UUID           // CreatedBy stores the created by value.
	CreatedAt   time.Time            // CreatedAt stores the created at value.
	UpdatedAt   time.Time            // UpdatedAt stores the updated at value.
	RetiredAt   *time.Time           // RetiredAt stores the retired at value.
	RevokedAt   *time.Time           // RevokedAt stores the revoked at value.
	Description string               // Description stores the description value.
}

// ThemePreviewToken grants short-lived preview access to one version.
type ThemePreviewToken struct {
	ID            uuid.UUID            // ID stores the i d value.
	VersionID     uuid.UUID            // VersionID stores the version i d value.
	TokenHash     string               // TokenHash stores the token hash value.
	PersonaKind   PreviewPersonaKind   // PersonaKind stores the persona kind value.
	PersonaSource PreviewPersonaSource // PersonaSource stores the persona source value.
	PersonaUserID *uuid.UUID           // PersonaUserID stores the persona user i d value.
	ExpiresAt     time.Time            // ExpiresAt stores the expires at value.
	CreatedBy     *uuid.UUID           // CreatedBy stores the created by value.
	CreatedAt     time.Time            // CreatedAt stores the created at value.
	RevokedAt     *time.Time           // RevokedAt stores the revoked at value.
}
