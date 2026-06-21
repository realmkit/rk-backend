package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	// defaultEnvFile is the default dotenv file read by Load.
	defaultEnvFile = ".env"

	// defaultPrefix is the environment variable prefix used by Load.
	defaultPrefix = "REALMKIT"
)

// Option changes the behavior of the configuration loader.
type Option func(*loader)

// loader contains configuration loading settings.
type loader struct {
	envFile string // envFile stores the env file value.
	prefix  string // prefix stores the prefix value.
}

// Load reads RealmKit configuration from defaults, .env, and environment variables.
func Load(options ...Option) (Config, error) {
	settings := loader{
		envFile: defaultEnvFile,
		prefix:  defaultPrefix,
	}

	for _, option := range options {
		option(&settings)
	}

	fields, err := schemaFor(Config{})
	if err != nil {
		return Config{}, err
	}

	source := newViper(settings.prefix)
	for _, field := range fields {
		if field.hasDefault {
			source.SetDefault(field.key, field.defaultValue)
		}
		if err := source.BindEnv(field.key, field.env(settings.prefix)); err != nil {
			return Config{}, fmt.Errorf("bind environment variable %s: %w", field.env(settings.prefix), err)
		}
	}

	if err := applyEnvFile(source, fields, settings); err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := source.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode configuration: %w", err)
	}

	if err := validateRequired(source, fields); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// WithEnvFile sets the .env file path used by the configuration loader.
func WithEnvFile(path string) Option {
	return func(settings *loader) {
		settings.envFile = path
	}
}

// applyEnvFile applies .env values as defaults before real environment overrides.
func applyEnvFile(source *viper.Viper, fields []fieldSpec, settings loader) error {
	if settings.envFile == "" {
		return nil
	}

	file := viper.New()
	file.SetConfigFile(settings.envFile)
	file.SetConfigType("env")

	if err := file.ReadInConfig(); err != nil {
		var missing viper.ConfigFileNotFoundError
		if errors.As(err, &missing) || os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read environment file %s: %w", settings.envFile, err)
	}

	for _, field := range fields {
		envName := field.env(settings.prefix)
		switch {
		case file.IsSet(envName):
			source.SetDefault(field.key, file.Get(envName))
		case file.IsSet(field.key):
			source.SetDefault(field.key, file.Get(field.key))
		}
	}

	return nil
}

// newViper creates a Viper instance configured for REALMKIT-style env keys.
func newViper(prefix string) *viper.Viper {
	source := viper.New()
	source.SetEnvPrefix(prefix)
	source.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	source.AutomaticEnv()
	return source
}

// validateRequired verifies fields without default tags are present and nonblank.
func validateRequired(source *viper.Viper, fields []fieldSpec) error {
	for _, field := range fields {
		if field.hasDefault {
			continue
		}
		if !source.IsSet(field.key) {
			return fmt.Errorf("missing required configuration %s", field.key)
		}
		if value, ok := source.Get(field.key).(string); ok && strings.TrimSpace(value) == "" {
			return fmt.Errorf("missing required configuration %s", field.key)
		}
	}

	return nil
}
