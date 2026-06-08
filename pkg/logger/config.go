package logger

// Config contains structured logging settings.
type Config struct {
	// Level is the minimum log severity emitted by the backend.
	Level string `mapstructure:"level" default:"info"`
}
