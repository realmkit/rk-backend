// Package postgres stores themes in PostgreSQL through GORM.
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// ThemeModel is the GORM model for theme families.
type ThemeModel struct {
	orm.ID
	Key             string `gorm:"not null;index"`
	Name            string `gorm:"not null"`
	Description     string `gorm:"not null;default:''"`
	Status          string `gorm:"not null;index"`
	CreatedByUserID *uuid.UUID
	UpdatedByUserID *uuid.UUID
	Version         uint64 `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (ThemeModel) TableName() string { return "themes" }

// VersionModel is the GORM model for theme versions.
type VersionModel struct {
	orm.ID
	ThemeID            uuid.UUID `gorm:"type:uuid;not null;index"`
	Semver             string    `gorm:"not null;default:''"`
	Label              string    `gorm:"not null;default:''"`
	Status             string    `gorm:"not null;index"`
	SourceKind         string    `gorm:"not null"`
	SourceReference    string    `gorm:"not null;default:''"`
	PackageStorageKey  string    `gorm:"not null;default:''"`
	PackageSizeBytes   int64     `gorm:"not null;default:0"`
	ManifestJSON       string    `gorm:"type:jsonb;not null;default:'{}'"`
	SettingsSchemaJSON string    `gorm:"type:jsonb;not null;default:'{}'"`
	SettingsDataJSON   string    `gorm:"type:jsonb;not null;default:'{}'"`
	IntegritySHA256    string    `gorm:"not null;default:'';index"`
	PublishedAt        *time.Time
	PublishedByUserID  *uuid.UUID
	CreatedByUserID    *uuid.UUID
	UpdatedByUserID    *uuid.UUID
	Version            uint64 `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (VersionModel) TableName() string { return "theme_versions" }

// FileModel is the GORM model for theme version files.
type FileModel struct {
	orm.ID
	VersionID         uuid.UUID `gorm:"type:uuid;not null;index"`
	Kind              string    `gorm:"not null;index"`
	Path              string    `gorm:"not null;index"`
	ContentSHA256     string    `gorm:"not null"`
	ContentStorageKey string    `gorm:"not null;default:''"`
	ContentText       string    `gorm:"not null;default:''"`
	SizeBytes         int64     `gorm:"not null;default:0"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (FileModel) TableName() string { return "theme_files" }

// AssetModel is the GORM model for theme version assets.
type AssetModel struct {
	orm.ID
	VersionID      uuid.UUID `gorm:"type:uuid;not null;index"`
	FileID         uuid.UUID `gorm:"type:uuid;not null;index"`
	Path           string    `gorm:"not null;index"`
	ContentType    string    `gorm:"not null"`
	SizeBytes      int64     `gorm:"not null;default:0"`
	ContentSHA256  string    `gorm:"not null"`
	StorageKey     string    `gorm:"not null;index"`
	PublicURL      string    `gorm:"not null;default:''"`
	IntegrityValue string    `gorm:"not null;default:''"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (AssetModel) TableName() string { return "theme_assets" }

// ActivationModel is the GORM model for active version pointers.
type ActivationModel struct {
	orm.ID
	ThemeID           uuid.UUID `gorm:"type:uuid;not null;index"`
	VersionID         uuid.UUID `gorm:"type:uuid;not null;index"`
	Environment       string    `gorm:"not null;index"`
	IsCurrent         bool      `gorm:"not null;default:true;index"`
	Reason            string    `gorm:"not null;default:''"`
	SettingsDataJSON  string    `gorm:"type:jsonb;not null;default:'{}'"`
	ActivatedByUserID *uuid.UUID
	ActivatedAt       time.Time `gorm:"not null;index"`
	CreatedAt         time.Time `gorm:"not null"`
	orm.SoftDelete
}

// TableName returns the database table name.
func (ActivationModel) TableName() string { return "theme_activations" }

// IssueModel is the GORM model for theme validation issues.
type IssueModel struct {
	orm.ID
	VersionID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Severity     string    `gorm:"not null;index"`
	Code         string    `gorm:"not null;index"`
	Path         string    `gorm:"not null;default:''"`
	Message      string    `gorm:"not null"`
	Line         int       `gorm:"not null;default:0"`
	ColumnNumber int       `gorm:"not null;default:0"`
	DetailsJSON  string    `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt    time.Time `gorm:"not null"`
	orm.SoftDelete
}

// TableName returns the database table name.
func (IssueModel) TableName() string { return "theme_validation_issues" }

// SignatureModel is the GORM model for package signatures.
type SignatureModel struct {
	orm.ID
	VersionID          uuid.UUID `gorm:"type:uuid;not null;index"`
	KeyID              string    `gorm:"not null;default:'';index"`
	Algorithm          string    `gorm:"not null"`
	VerificationStatus string    `gorm:"not null;index"`
	Signature          string    `gorm:"not null;default:''"`
	SignedManifestHash string    `gorm:"not null;default:''"`
	VerifiedAt         *time.Time
	CreatedAt          time.Time `gorm:"not null"`
	orm.SoftDelete
}

// TableName returns the database table name.
func (SignatureModel) TableName() string { return "theme_package_signatures" }

// SigningKeyModel is the GORM model for trusted package signing keys.
type SigningKeyModel struct {
	orm.ID
	KeyID           string `gorm:"not null;index"`
	Algorithm       string `gorm:"not null"`
	PublicKey       string `gorm:"not null"`
	TrustLevel      string `gorm:"not null;index"`
	Status          string `gorm:"not null;index"`
	Source          string `gorm:"not null"`
	NotBefore       *time.Time
	NotAfter        *time.Time
	CreatedByUserID *uuid.UUID
	Description     string `gorm:"not null;default:''"`
	orm.Timestamps
	RetiredAt *time.Time
	RevokedAt *time.Time
	orm.SoftDelete
}

// TableName returns the database table name.
func (SigningKeyModel) TableName() string { return "theme_signing_keys" }

// PreviewTokenModel is the GORM model for preview tokens.
type PreviewTokenModel struct {
	orm.ID
	VersionID       uuid.UUID `gorm:"type:uuid;not null;index"`
	TokenHash       string    `gorm:"not null;index"`
	PersonaKind     string    `gorm:"not null"`
	PersonaSource   string    `gorm:"not null"`
	PersonaUserID   *uuid.UUID
	ExpiresAt       time.Time `gorm:"not null;index"`
	CreatedByUserID *uuid.UUID
	CreatedAt       time.Time `gorm:"not null"`
	RevokedAt       *time.Time
	orm.SoftDelete
}

// TableName returns the database table name.
func (PreviewTokenModel) TableName() string { return "theme_preview_tokens" }
