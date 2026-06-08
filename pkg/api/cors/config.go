package cors

// Config contains browser cross-origin settings.
type Config struct {
	// Enabled reports whether CORS middleware is active.
	Enabled bool `mapstructure:"enabled" default:"true"`

	// AllowOrigins contains the comma-separated allowed browser origins.
	AllowOrigins string `mapstructure:"allow_origins" default:"http://localhost:3000,http://127.0.0.1:3000"`
}
