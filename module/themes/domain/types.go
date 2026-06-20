// Package domain owns theme contract types and pure validation catalogs.
package domain

// ThemeStatus is the lifecycle state of a theme family.
type ThemeStatus string

// VersionStatus is the lifecycle state of one immutable theme version.
type VersionStatus string

// SourceKind identifies how a theme version entered RealmKit.
type SourceKind string

// FileKind identifies a logical file inside a theme version.
type FileKind string

// ActivationEnvironment identifies the public or preview active theme pointer.
type ActivationEnvironment string

// SignatureAlgorithm identifies a supported package signature algorithm.
type SignatureAlgorithm string

// SignatureVerificationStatus is the outcome of package signature verification.
type SignatureVerificationStatus string

// SigningKeyTrustLevel identifies who owns a trusted signing key.
type SigningKeyTrustLevel string

// SigningKeyStatus is the lifecycle state of a trusted signing key.
type SigningKeyStatus string

// SigningKeySource identifies whether a key was configured or persisted.
type SigningKeySource string

// ValidationSeverity identifies how strongly a validation issue blocks work.
type ValidationSeverity string

// ValidationIssueCode identifies one structured theme validation diagnostic.
type ValidationIssueCode string

// RouteKind identifies one public route-data contract.
type RouteKind string

// RichTextProfile identifies one sanitizer policy.
type RichTextProfile string

// PreviewPersonaKind identifies the viewer lens used by theme preview.
type PreviewPersonaKind string

// PreviewPersonaSource identifies whether preview data is synthetic or real.
type PreviewPersonaSource string

// IdempotencyRequirement documents retry semantics for one command.
type IdempotencyRequirement string

const (
	// ThemeStatusDraft is editable before public availability.
	ThemeStatusDraft ThemeStatus = "draft"
	// ThemeStatusAvailable can receive activations.
	ThemeStatusAvailable ThemeStatus = "available"
	// ThemeStatusArchived is retained for history.
	ThemeStatusArchived ThemeStatus = "archived"
)

const (
	// VersionStatusDraft can be edited.
	VersionStatusDraft VersionStatus = "draft"
	// VersionStatusValidating is currently being checked.
	VersionStatusValidating VersionStatus = "validating"
	// VersionStatusValid passed validation.
	VersionStatusValid VersionStatus = "valid"
	// VersionStatusInvalid failed validation.
	VersionStatusInvalid VersionStatus = "invalid"
	// VersionStatusPublished was activated publicly at least once.
	VersionStatusPublished VersionStatus = "published"
	// VersionStatusArchived is retained but no longer publishable.
	VersionStatusArchived VersionStatus = "archived"
)

const (
	// SourceUpload came from a zip upload.
	SourceUpload SourceKind = "upload"
	// SourceEditor came from the browser editor.
	SourceEditor SourceKind = "editor"
	// SourceGit came from a future git import.
	SourceGit SourceKind = "git"
	// SourceSystem came from a bundled RealmKit theme.
	SourceSystem SourceKind = "system"
)

const (
	// FileKindLayout is a layout Liquid file.
	FileKindLayout FileKind = "layout"
	// FileKindTemplate is a route template Liquid file.
	FileKindTemplate FileKind = "template"
	// FileKindSection is a section Liquid file.
	FileKindSection FileKind = "section"
	// FileKindSnippet is a snippet Liquid file.
	FileKindSnippet FileKind = "snippet"
	// FileKindAsset is a versioned static asset.
	FileKindAsset FileKind = "asset"
	// FileKindConfig is a theme config JSON file.
	FileKindConfig FileKind = "config"
	// FileKindLocale is a translation JSON file.
	FileKindLocale FileKind = "locale"
	// FileKindManifest is the package manifest.
	FileKindManifest FileKind = "manifest"
	// FileKindSignature is the detached signature envelope.
	FileKindSignature FileKind = "signature"
)

const (
	// EnvironmentPublic is the active public theme pointer.
	EnvironmentPublic ActivationEnvironment = "public"
	// EnvironmentPreview is the active preview theme pointer.
	EnvironmentPreview ActivationEnvironment = "preview"
)

const (
	// SignatureAlgorithmEd25519 verifies detached Ed25519 package signatures.
	SignatureAlgorithmEd25519 SignatureAlgorithm = "ed25519"
)

const (
	// SignatureMissing means no package signature was provided.
	SignatureMissing SignatureVerificationStatus = "missing"
	// SignatureVerified means a trusted key verified the package.
	SignatureVerified SignatureVerificationStatus = "verified"
	// SignatureUntrusted means the key is unknown.
	SignatureUntrusted SignatureVerificationStatus = "untrusted"
	// SignatureRetired means the key can validate only older signatures.
	SignatureRetired SignatureVerificationStatus = "retired"
	// SignatureRevoked means the key can no longer validate activation.
	SignatureRevoked SignatureVerificationStatus = "revoked"
	// SignatureInvalid means the signature or manifest hash failed verification.
	SignatureInvalid SignatureVerificationStatus = "invalid"
)

const (
	// TrustLevelSystem identifies built-in RealmKit keys.
	TrustLevelSystem SigningKeyTrustLevel = "system"
	// TrustLevelVendor identifies vendor or marketplace keys.
	TrustLevelVendor SigningKeyTrustLevel = "vendor"
	// TrustLevelOperator identifies operator-managed keys.
	TrustLevelOperator SigningKeyTrustLevel = "operator"
)

const (
	// SigningKeyTrusted can validate old and new packages.
	SigningKeyTrusted SigningKeyStatus = "trusted"
	// SigningKeyRetired validates packages signed before retirement.
	SigningKeyRetired SigningKeyStatus = "retired"
	// SigningKeyRevoked blocks new activations.
	SigningKeyRevoked SigningKeyStatus = "revoked"
)

const (
	// SigningKeySourceEnvironment came from runtime configuration.
	SigningKeySourceEnvironment SigningKeySource = "environment"
	// SigningKeySourceDatabase came from operator persistence.
	SigningKeySourceDatabase SigningKeySource = "database"
)

const (
	// SeverityError blocks activation.
	SeverityError ValidationSeverity = "error"
	// SeverityWarning requires operator attention.
	SeverityWarning ValidationSeverity = "warning"
	// SeverityInfo is informational.
	SeverityInfo ValidationSeverity = "info"
)

const (
	// IdempotencyRequired means the command must include Idempotency-Key.
	IdempotencyRequired IdempotencyRequirement = "required"
	// IdempotencyOptional means retries may include Idempotency-Key.
	IdempotencyOptional IdempotencyRequirement = "optional"
	// IdempotencyUnsupported means the route must not accept Idempotency-Key.
	IdempotencyUnsupported IdempotencyRequirement = "unsupported"
)
