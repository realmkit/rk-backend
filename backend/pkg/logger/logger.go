package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Option changes the behavior of the logger builder.
type Option func(*settings)

type settings struct {
	output      zapcore.WriteSyncer
	errorOutput zapcore.WriteSyncer
}

// New creates a JSON Zap logger for the supplied logging configuration.
func New(cfg Config, options ...Option) (*zap.Logger, error) {
	level, err := levelFor(cfg.Level)
	if err != nil {
		return nil, err
	}

	settings := settings{
		output:      zapcore.Lock(os.Stdout),
		errorOutput: zapcore.Lock(os.Stderr),
	}
	for _, option := range options {
		option(&settings)
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig()), settings.output, level)
	return zap.New(core, zap.ErrorOutput(settings.errorOutput), zap.AddCaller()), nil
}

// WithOutput sets the writer used for structured log entries.
func WithOutput(writer io.Writer) Option {
	return func(settings *settings) {
		settings.output = zapcore.AddSync(writer)
	}
}

// WithErrorOutput sets the writer used for Zap internal errors.
func WithErrorOutput(writer io.Writer) Option {
	return func(settings *settings) {
		settings.errorOutput = zapcore.AddSync(writer)
	}
}

func levelFor(value string) (zapcore.Level, error) {
	var level zapcore.Level
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		normalized = "info"
	}
	if err := level.UnmarshalText([]byte(normalized)); err != nil {
		return level, fmt.Errorf("parse log level %q: %w", value, err)
	}
	return level, nil
}

func encoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}
