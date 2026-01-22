// Package lad provides small, opinionated helpers for building zap loggers.
//
// It follows two principles:
//
//  1. New builds a logger without touching zap globals.
//  2. InitGlobal is explicit about side effects (it replaces zap's global logger).
//
// File output uses lumberjack for log rotation.
package lad

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// type alias
type (
	Logger        = zap.Logger
	SugaredLogger = zap.SugaredLogger
)

// L returns the current global zap Logger (zap.L()).
func L() *Logger { return zap.L() }

// S returns the current global zap SugaredLogger (zap.S()).
func S() *SugaredLogger { return zap.S() }

// DefaultTimeFormat is the default timestamp layout used by encoders.
const DefaultTimeFormat = "2006-01-02 15:04:05.000"

// Option configures logger building behavior.
type Option func(*config) error

type config struct {
	cores   []zapcore.Core
	zapOpts []zap.Option
}

// WithZapOptions appends raw zap options to the logger being built.
func WithZapOptions(opts ...zap.Option) Option {
	return func(c *config) error {
		c.zapOpts = append(c.zapOpts, opts...)
		return nil
	}
}

// WithCaller enables caller annotations.
func WithCaller() Option {
	return func(c *config) error {
		c.zapOpts = append(c.zapOpts, zap.AddCaller())
		return nil
	}
}

// WithCallerSkip sets the number of stack frames to skip when reporting caller.
// Useful when you wrap this logger behind another logging facade.
func WithCallerSkip(skip int) Option {
	return func(c *config) error {
		if skip < 0 {
			return errors.New("lad: caller skip must be >= 0")
		}
		c.zapOpts = append(c.zapOpts, zap.AddCallerSkip(skip))
		return nil
	}
}

// WithStacktrace enables stack traces at and above the given level.
func WithStacktrace(level zapcore.Level) Option {
	return func(c *config) error {
		c.zapOpts = append(c.zapOpts, zap.AddStacktrace(level))
		return nil
	}
}

// ConsoleConfig controls console output.
type ConsoleConfig struct {
	Level      zapcore.Level
	Colored    bool
	TimeFormat string   // Defaults to DefaultTimeFormat when empty.
	Output     *os.File // Defaults to os.Stdout when nil.
}

// WithConsole adds a console core to the logger.
func WithConsole(cc ConsoleConfig) Option {
	return func(c *config) error {
		out := cc.Output
		if out == nil {
			out = os.Stdout
		}

		encCfg := zap.NewProductionEncoderConfig()
		encCfg.EncodeTime = timeEncoder(orDefault(cc.TimeFormat, DefaultTimeFormat))

		if cc.Colored {
			encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		} else {
			encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		}

		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encCfg),
			zapcore.AddSync(out),
			cc.Level,
		)
		c.cores = append(c.cores, core)
		return nil
	}
}

// FileEncoding controls how file logs are encoded.
type FileEncoding string

const (
	// JSONEncoding writes logs as JSON (recommended for file output).
	JSONEncoding FileEncoding = "json"
	// ConsoleEncoding writes logs in console style (human-readable).
	ConsoleEncoding FileEncoding = "console"
)

// FileConfig controls rotating file output (powered by lumberjack).
type FileConfig struct {
	Level      zapcore.Level
	Filename   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool

	Encoding   FileEncoding // Defaults to JSONEncoding when empty.
	TimeFormat string       // Defaults to DefaultTimeFormat when empty.
}

// WithFile adds a rotating file core to the logger.
func WithFile(fc FileConfig) Option {
	return func(c *config) error {
		if strings.TrimSpace(fc.Filename) == "" {
			return errors.New("lad: FileConfig.Filename is required")
		}

		maxSize := fc.MaxSizeMB
		if maxSize <= 0 {
			maxSize = 100
		}

		encoding := fc.Encoding
		if encoding == "" {
			encoding = JSONEncoding
		}

		hook := &lumberjack.Logger{
			Filename:   fc.Filename,
			MaxSize:    maxSize,
			MaxBackups: fc.MaxBackups,
			MaxAge:     fc.MaxAgeDays,
			Compress:   fc.Compress,
		}

		encCfg := zap.NewProductionEncoderConfig()
		encCfg.EncodeTime = timeEncoder(orDefault(fc.TimeFormat, DefaultTimeFormat))
		encCfg.EncodeLevel = zapcore.CapitalLevelEncoder

		var enc zapcore.Encoder
		switch encoding {
		case JSONEncoding:
			enc = zapcore.NewJSONEncoder(encCfg)
		case ConsoleEncoding:
			enc = zapcore.NewConsoleEncoder(encCfg)
		default:
			return fmt.Errorf("lad: unknown FileEncoding %q", encoding)
		}

		core := zapcore.NewCore(
			enc,
			zapcore.AddSync(hook),
			fc.Level,
		)
		c.cores = append(c.cores, core)
		return nil
	}
}

// New builds a zap Logger with the given options.
// It does not modify zap's global logger.
func New(opts ...Option) (*Logger, error) {
	cfg := &config{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	// Default core if none provided: colored console at Debug level.
	if len(cfg.cores) == 0 {
		_ = WithConsole(ConsoleConfig{
			Level:      zap.DebugLevel,
			Colored:    true,
			TimeFormat: DefaultTimeFormat,
			Output:     os.Stdout,
		})(cfg)
	}

	core := zapcore.NewTee(cfg.cores...)
	return zap.New(core, cfg.zapOpts...), nil
}

// MustNew is like New but panics on error.
// This is intended for application startup code (e.g., main()).
func MustNew(opts ...Option) *Logger {
	l, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return l
}

// InitGlobal builds a logger and replaces zap's global logger (zap.ReplaceGlobals).
func InitGlobal(opts ...Option) error {
	l, err := New(opts...)
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(l)
	return nil
}

// MustInitGlobal is like InitGlobal but panics on error.
// This is intended for application startup code (e.g., main()).
func MustInitGlobal(opts ...Option) {
	err := InitGlobal(opts...)
	if err != nil {
		panic(err)
	}
}

// RedirectStdLog redirects the standard library's package-global logger (log.Print, etc.)
// to the supplied zap logger at InfoLevel.
//
// It returns a function that restores the original standard logger configuration.
func RedirectStdLog(l *Logger) func() {
	return zap.RedirectStdLog(l)
}

// RedirectStdLogAt is like RedirectStdLog but allows specifying the level.
func RedirectStdLogAt(l *Logger, level zapcore.Level) (func(), error) {
	return zap.RedirectStdLogAt(l, level)
}

// Sync flushes any buffered log entries.
// It ignores common errors produced when syncing stdout/stderr in some environments.
func Sync(l *Logger) error {
	if l == nil {
		return nil
	}
	err := l.Sync()
	if err == nil {
		return nil
	}
	if isIgnorableSyncErr(err) {
		return nil
	}
	return err
}

func timeEncoder(layout string) func(time.Time, zapcore.PrimitiveArrayEncoder) {
	return func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
		pae.AppendString(t.Format(layout))
	}
}

func orDefault(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func isIgnorableSyncErr(err error) bool {
	// Common in containers/terminals: "sync /dev/stderr: invalid argument"
	// Also happens on Windows; stdout/stderr sync is often unsupported there.
	if runtime.GOOS == "windows" {
		return true
	}
	if errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY) || errors.Is(err, syscall.EBADF) {
		return true
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "invalid argument") || strings.Contains(msg, "inappropriate ioctl") {
		return true
	}
	return false
}
