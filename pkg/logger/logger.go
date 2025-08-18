// Package logger provides structured logging for gearbox applications.
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Level represents logging levels
type Level int8

const (
	// DebugLevel defines debug log level.
	DebugLevel Level = iota
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel
	// NoLevel defines an absent log level.
	NoLevel
	// Disabled disables the logger.
	Disabled
)

// Logger wraps zerolog.Logger with gearbox-specific functionality.
type Logger struct {
	logger zerolog.Logger
}

// Config represents logger configuration.
type Config struct {
	Level      Level
	Pretty     bool
	TimeFormat string
	Output     io.Writer
}

// DefaultConfig returns a default logger configuration.
func DefaultConfig() Config {
	return Config{
		Level:      InfoLevel,
		Pretty:     true,
		TimeFormat: time.RFC3339,
		Output:     os.Stdout,
	}
}

// New creates a new structured logger with the given configuration.
func New(config Config) *Logger {
	// Convert our Level to zerolog.Level
	zeroLevel := zerolog.Level(config.Level)
	
	// Configure zerolog global settings
	zerolog.SetGlobalLevel(zeroLevel)
	
	var output io.Writer = config.Output
	if config.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        config.Output,
			TimeFormat: config.TimeFormat,
			FormatLevel: func(i interface{}) string {
				level := ""
				if ll, ok := i.(string); ok {
					switch ll {
					case "debug":
						level = "🐛 DEBUG"
					case "info":
						level = "📋 INFO "
					case "warn":
						level = "⚠️  WARN "
					case "error":
						level = "❌ ERROR"
					case "fatal":
						level = "💀 FATAL"
					case "panic":
						level = "🚨 PANIC"
					default:
						level = "📝 " + ll
					}
				}
				return level
			},
			FormatCaller: func(i interface{}) string {
				if caller, ok := i.(string); ok {
					return "(" + caller + ")"
				}
				return ""
			},
		}
	}
	
	logger := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()
	
	return &Logger{logger: logger}
}

// NewDefault creates a logger with default configuration.
func NewDefault() *Logger {
	return New(DefaultConfig())
}

// NewQuiet creates a quiet logger (errors only).
func NewQuiet() *Logger {
	config := DefaultConfig()
	config.Level = ErrorLevel
	config.Pretty = false
	return New(config)
}

// NewVerbose creates a verbose logger (debug level).
func NewVerbose() *Logger {
	config := DefaultConfig()
	config.Level = DebugLevel
	return New(config)
}

// Debug logs a message at debug level.
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted message at debug level.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info logs a message at info level.
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted message at info level.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn logs a message at warn level.
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted message at warn level.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error logs a message at error level.
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted message at error level.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// ErrorWithErr logs an error with additional context.
func (l *Logger) ErrorWithErr(err error, msg string) {
	l.logger.Error().Err(err).Msg(msg)
}

// Fatal logs a message at fatal level and exits.
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted message at fatal level and exits.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// WithField adds a field to log entries.
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{logger: l.logger.With().Interface(key, value).Logger()}
}

// WithFields adds multiple fields to log entries.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &Logger{logger: ctx.Logger()}
}

// WithError adds an error field to log entries.
func (l *Logger) WithError(err error) *Logger {
	return &Logger{logger: l.logger.With().Err(err).Logger()}
}

// Progress logs progress information for long-running operations.
func (l *Logger) Progress(operation string, current, total int) {
	percentage := float64(current) / float64(total) * 100
	l.logger.Info().
		Str("operation", operation).
		Int("current", current).
		Int("total", total).
		Float64("percentage", percentage).
		Msgf("Progress: %s (%d/%d, %.1f%%)", operation, current, total, percentage)
}

// Operation creates a logger scoped to a specific operation.
func (l *Logger) Operation(operation string) *Logger {
	return l.WithField("operation", operation)
}

// Tool creates a logger scoped to a specific tool.
func (l *Logger) Tool(tool string) *Logger {
	return l.WithField("tool", tool)
}

// Duration logs operation duration.
func (l *Logger) Duration(operation string, duration time.Duration) {
	l.logger.Info().
		Str("operation", operation).
		Dur("duration", duration).
		Msgf("Operation completed: %s (took %v)", operation, duration)
}

// Global logger instance
var globalLogger *Logger

func init() {
	globalLogger = NewDefault()
}

// SetGlobalLogger sets the global logger instance.
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance.
func GetGlobalLogger() *Logger {
	return globalLogger
}

// Global logging functions using the global logger

// Debug logs a debug message using the global logger.
func Debug(msg string) {
	globalLogger.Debug(msg)
}

// Debugf logs a formatted debug message using the global logger.
func Debugf(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

// Info logs an info message using the global logger.
func Info(msg string) {
	globalLogger.Info(msg)
}

// Infof logs a formatted info message using the global logger.
func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// Warn logs a warning message using the global logger.
func Warn(msg string) {
	globalLogger.Warn(msg)
}

// Warnf logs a formatted warning message using the global logger.
func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

// Error logs an error message using the global logger.
func Error(msg string) {
	globalLogger.Error(msg)
}

// Errorf logs a formatted error message using the global logger.
func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

// ErrorWithErr logs an error with additional context using the global logger.
func ErrorWithErr(err error, msg string) {
	globalLogger.ErrorWithErr(err, msg)
}

// Fatal logs a fatal message and exits using the global logger.
func Fatal(msg string) {
	globalLogger.Fatal(msg)
}

// Fatalf logs a formatted fatal message and exits using the global logger.
func Fatalf(format string, args ...interface{}) {
	globalLogger.Fatalf(format, args...)
}