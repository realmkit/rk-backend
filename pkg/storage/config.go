package storage

// Config contains S3-compatible object storage settings.
type Config struct {
	// Bucket is the S3 bucket where GameHub stores assets.
	Bucket string `mapstructure:"bucket"`

	// Region is the S3 region.
	Region string `mapstructure:"region" default:"auto"`

	// Endpoint is the S3-compatible service endpoint.
	Endpoint string `mapstructure:"endpoint"`

	// AccessKeyID is the S3 access key identifier.
	AccessKeyID string `mapstructure:"access_key_id"`

	// SecretAccessKey is the S3 secret access key.
	SecretAccessKey string `mapstructure:"secret_access_key"`

	// PublicBaseURL is an optional CDN or public bucket base URL.
	PublicBaseURL string `mapstructure:"public_base_url" default:""`
}
