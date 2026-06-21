// Package postgres stores themes in PostgreSQL through GORM.
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// ThemeModel is the GORM model for theme families.
type ThemeModel struct {
	orm.ID                     // ID embeds shared fields.
	Key             string     `gorm:"not null;index"`      // Key stores the key value.
	Name            string     `gorm:"not null"`            // Name stores the name value.
	Description     string     `gorm:"not null;default:''"` // Description stores the description value.
	Status          string     `gorm:"not null;index"`      // Status stores the status value.
	CreatedByUserID *uuid.UUID // CreatedByUserID stores the created by user i d value.
	UpdatedByUserID *uuid.UUID // UpdatedByUserID stores the updated by user i d value.
	Version         uint64     `gorm:"not null;default:1"` // Version stores the version value.
	orm.Timestamps             // Timestamps embeds shared fields.
	orm.SoftDelete             // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (ThemeModel) TableName() string { return "themes" }

// VersionModel is the GORM model for theme versions.
type VersionModel struct {
	orm.ID                        // ID embeds shared fields.
	ThemeID            uuid.UUID  `gorm:"type:uuid;not null;index"`         // ThemeID stores the theme i d value.
	Semver             string     `gorm:"not null;default:''"`              // Semver stores the semver value.
	Label              string     `gorm:"not null;default:''"`              // Label stores the label value.
	Status             string     `gorm:"not null;index"`                   // Status stores the status value.
	SourceKind         string     `gorm:"not null"`                         // SourceKind stores the source kind value.
	SourceReference    string     `gorm:"not null;default:''"`              // SourceReference stores the source reference value.
	PackageStorageKey  string     `gorm:"not null;default:''"`              // PackageStorageKey stores the package storage key value.
	PackageSizeBytes   int64      `gorm:"not null;default:0"`               // PackageSizeBytes stores the package size bytes value.
	ManifestJSON       string     `gorm:"type:jsonb;not null;default:'{}'"` // ManifestJSON stores the manifest j s o n value.
	SettingsSchemaJSON string     `gorm:"type:jsonb;not null;default:'{}'"` // SettingsSchemaJSON stores the settings schema j s o n value.
	SettingsDataJSON   string     `gorm:"type:jsonb;not null;default:'{}'"` // SettingsDataJSON stores the settings data j s o n value.
	IntegritySHA256    string     `gorm:"not null;default:'';index"`        // IntegritySHA256 stores the integrity s h a256 value.
	PublishedAt        *time.Time // PublishedAt stores the published at value.
	PublishedByUserID  *uuid.UUID // PublishedByUserID stores the published by user i d value.
	CreatedByUserID    *uuid.UUID // CreatedByUserID stores the created by user i d value.
	UpdatedByUserID    *uuid.UUID // UpdatedByUserID stores the updated by user i d value.
	Version            uint64     `gorm:"not null;default:1"` // Version stores the version value.
	orm.Timestamps                // Timestamps embeds shared fields.
	orm.SoftDelete                // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (VersionModel) TableName() string { return "theme_versions" }

// FileModel is the GORM model for theme version files.
type FileModel struct {
	orm.ID                      // ID embeds shared fields.
	VersionID         uuid.UUID `gorm:"type:uuid;not null;index"` // VersionID stores the version i d value.
	Kind              string    `gorm:"not null;index"`           // Kind stores the kind value.
	Path              string    `gorm:"not null;index"`           // Path stores the path value.
	ContentSHA256     string    `gorm:"not null"`                 // ContentSHA256 stores the content s h a256 value.
	ContentStorageKey string    `gorm:"not null;default:''"`      // ContentStorageKey stores the content storage key value.
	ContentText       string    `gorm:"not null;default:''"`      // ContentText stores the content text value.
	SizeBytes         int64     `gorm:"not null;default:0"`       // SizeBytes stores the size bytes value.
	orm.Timestamps              // Timestamps embeds shared fields.
	orm.SoftDelete              // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (FileModel) TableName() string { return "theme_files" }

// AssetModel is the GORM model for theme version assets.
type AssetModel struct {
	orm.ID                   // ID embeds shared fields.
	VersionID      uuid.UUID `gorm:"type:uuid;not null;index"` // VersionID stores the version i d value.
	FileID         uuid.UUID `gorm:"type:uuid;not null;index"` // FileID stores the file i d value.
	Path           string    `gorm:"not null;index"`           // Path stores the path value.
	ContentType    string    `gorm:"not null"`                 // ContentType stores the content type value.
	SizeBytes      int64     `gorm:"not null;default:0"`       // SizeBytes stores the size bytes value.
	ContentSHA256  string    `gorm:"not null"`                 // ContentSHA256 stores the content s h a256 value.
	StorageKey     string    `gorm:"not null;index"`           // StorageKey stores the storage key value.
	PublicURL      string    `gorm:"not null;default:''"`      // PublicURL stores the public u r l value.
	IntegrityValue string    `gorm:"not null;default:''"`      // IntegrityValue stores the integrity value value.
	orm.Timestamps           // Timestamps embeds shared fields.
	orm.SoftDelete           // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (AssetModel) TableName() string { return "theme_assets" }

// ActivationModel is the GORM model for active version pointers.
type ActivationModel struct {
	orm.ID                       // ID embeds shared fields.
	ThemeID           uuid.UUID  `gorm:"type:uuid;not null;index"`         // ThemeID stores the theme i d value.
	VersionID         uuid.UUID  `gorm:"type:uuid;not null;index"`         // VersionID stores the version i d value.
	Environment       string     `gorm:"not null;index"`                   // Environment stores the environment value.
	IsCurrent         bool       `gorm:"not null;default:true;index"`      // IsCurrent stores the is current value.
	Reason            string     `gorm:"not null;default:''"`              // Reason stores the reason value.
	SettingsDataJSON  string     `gorm:"type:jsonb;not null;default:'{}'"` // SettingsDataJSON stores the settings data j s o n value.
	ActivatedByUserID *uuid.UUID // ActivatedByUserID stores the activated by user i d value.
	ActivatedAt       time.Time  `gorm:"not null;index"` // ActivatedAt stores the activated at value.
	CreatedAt         time.Time  `gorm:"not null"`       // CreatedAt stores the created at value.
	orm.SoftDelete               // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (ActivationModel) TableName() string { return "theme_activations" }

// IssueModel is the GORM model for theme validation issues.
type IssueModel struct {
	orm.ID                   // ID embeds shared fields.
	VersionID      uuid.UUID `gorm:"type:uuid;not null;index"`         // VersionID stores the version i d value.
	Severity       string    `gorm:"not null;index"`                   // Severity stores the severity value.
	Code           string    `gorm:"not null;index"`                   // Code stores the code value.
	Path           string    `gorm:"not null;default:''"`              // Path stores the path value.
	Message        string    `gorm:"not null"`                         // Message stores the message value.
	Line           int       `gorm:"not null;default:0"`               // Line stores the line value.
	ColumnNumber   int       `gorm:"not null;default:0"`               // ColumnNumber stores the column number value.
	DetailsJSON    string    `gorm:"type:jsonb;not null;default:'{}'"` // DetailsJSON stores the details j s o n value.
	CreatedAt      time.Time `gorm:"not null"`                         // CreatedAt stores the created at value.
	orm.SoftDelete           // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (IssueModel) TableName() string { return "theme_validation_issues" }

// SignatureModel is the GORM model for package signatures.
type SignatureModel struct {
	orm.ID                        // ID embeds shared fields.
	VersionID          uuid.UUID  `gorm:"type:uuid;not null;index"`  // VersionID stores the version i d value.
	KeyID              string     `gorm:"not null;default:'';index"` // KeyID stores the key i d value.
	Algorithm          string     `gorm:"not null"`                  // Algorithm stores the algorithm value.
	VerificationStatus string     `gorm:"not null;index"`            // VerificationStatus stores the verification status value.
	Signature          string     `gorm:"not null;default:''"`       // Signature stores the signature value.
	SignedManifestHash string     `gorm:"not null;default:''"`       // SignedManifestHash stores the signed manifest hash value.
	VerifiedAt         *time.Time // VerifiedAt stores the verified at value.
	CreatedAt          time.Time  `gorm:"not null"` // CreatedAt stores the created at value.
	orm.SoftDelete                // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (SignatureModel) TableName() string { return "theme_package_signatures" }

// SigningKeyModel is the GORM model for trusted package signing keys.
type SigningKeyModel struct {
	orm.ID                     // ID embeds shared fields.
	KeyID           string     `gorm:"not null;index"` // KeyID stores the key i d value.
	Algorithm       string     `gorm:"not null"`       // Algorithm stores the algorithm value.
	PublicKey       string     `gorm:"not null"`       // PublicKey stores the public key value.
	TrustLevel      string     `gorm:"not null;index"` // TrustLevel stores the trust level value.
	Status          string     `gorm:"not null;index"` // Status stores the status value.
	Source          string     `gorm:"not null"`       // Source stores the source value.
	NotBefore       *time.Time // NotBefore stores the not before value.
	NotAfter        *time.Time // NotAfter stores the not after value.
	CreatedByUserID *uuid.UUID // CreatedByUserID stores the created by user i d value.
	Description     string     `gorm:"not null;default:''"` // Description stores the description value.
	orm.Timestamps             // Timestamps embeds shared fields.
	RetiredAt       *time.Time // RetiredAt stores the retired at value.
	RevokedAt       *time.Time // RevokedAt stores the revoked at value.
	orm.SoftDelete             // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (SigningKeyModel) TableName() string { return "theme_signing_keys" }

// PreviewTokenModel is the GORM model for preview tokens.
type PreviewTokenModel struct {
	orm.ID                     // ID embeds shared fields.
	VersionID       uuid.UUID  `gorm:"type:uuid;not null;index"` // VersionID stores the version i d value.
	TokenHash       string     `gorm:"not null;index"`           // TokenHash stores the token hash value.
	PersonaKind     string     `gorm:"not null"`                 // PersonaKind stores the persona kind value.
	PersonaSource   string     `gorm:"not null"`                 // PersonaSource stores the persona source value.
	PersonaUserID   *uuid.UUID // PersonaUserID stores the persona user i d value.
	ExpiresAt       time.Time  `gorm:"not null;index"` // ExpiresAt stores the expires at value.
	CreatedByUserID *uuid.UUID // CreatedByUserID stores the created by user i d value.
	CreatedAt       time.Time  `gorm:"not null"` // CreatedAt stores the created at value.
	RevokedAt       *time.Time // RevokedAt stores the revoked at value.
	orm.SoftDelete             // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (PreviewTokenModel) TableName() string { return "theme_preview_tokens" }
