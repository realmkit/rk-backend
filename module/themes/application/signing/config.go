package signing

// Config contains theme signature verification policy.
type Config struct {
	AllowUnsignedPackages bool `mapstructure:"allow_unsigned_packages" default:"false"` // AllowUnsignedPackages stores the allow unsigned packages value.
}
