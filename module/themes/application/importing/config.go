package importing

// Config contains theme package import limits.
type Config struct {
	MaxPackageBytes     int64   `mapstructure:"max_package_bytes" default:"26214400"`
	MaxExtractedBytes   int64   `mapstructure:"max_extracted_bytes" default:"104857600"`
	MaxFileCount        int     `mapstructure:"max_file_count" default:"1000"`
	MaxTextFileBytes    int64   `mapstructure:"max_text_file_bytes" default:"1048576"`
	MaxCompressionRatio float64 `mapstructure:"max_compression_ratio" default:"100"`
	StoragePrefix       string  `mapstructure:"storage_prefix" default:"themes"`
}

const (
	// DefaultMaxPackageBytes is the first-version zip upload cap.
	DefaultMaxPackageBytes int64 = 25 * 1024 * 1024
	// DefaultMaxExtractedBytes is the first-version extracted package cap.
	DefaultMaxExtractedBytes int64 = 100 * 1024 * 1024
	// DefaultMaxFileCount is the first-version package file count cap.
	DefaultMaxFileCount = 1000
	// DefaultMaxTextFileBytes is the browser-editable text file cap.
	DefaultMaxTextFileBytes int64 = 1024 * 1024
	// DefaultMaxCompressionRatio is the zip bomb ratio guard.
	DefaultMaxCompressionRatio = 100
)

// Defaults returns config with explicit default values applied.
func (cfg Config) Defaults() Config {
	if cfg.MaxPackageBytes <= 0 {
		cfg.MaxPackageBytes = DefaultMaxPackageBytes
	}
	if cfg.MaxExtractedBytes <= 0 {
		cfg.MaxExtractedBytes = DefaultMaxExtractedBytes
	}
	if cfg.MaxFileCount <= 0 {
		cfg.MaxFileCount = DefaultMaxFileCount
	}
	if cfg.MaxTextFileBytes <= 0 {
		cfg.MaxTextFileBytes = DefaultMaxTextFileBytes
	}
	if cfg.MaxCompressionRatio <= 0 {
		cfg.MaxCompressionRatio = DefaultMaxCompressionRatio
	}
	if cfg.StoragePrefix == "" {
		cfg.StoragePrefix = "themes"
	}
	return cfg
}
