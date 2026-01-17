package lad_test

import (
	"testing"

	"github.com/omivix/lad"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestMustInitGlobalConsole(t *testing.T) {
	lad.MustInitGlobal(
		lad.WithConsole(lad.ConsoleConfig{
			Level:   zap.DebugLevel,
			Colored: true,
		}),
		lad.WithCaller(),
		lad.WithStacktrace(zapcore.ErrorLevel),
	)

	defer func() { _ = lad.Sync(lad.L()) }()

	lad.L().Info("service started")
}

func TestMustInitGlobalFile(t *testing.T) {
	lad.MustInitGlobal(
		lad.WithFile(lad.FileConfig{
			Level:      zap.InfoLevel,
			Filename:   "./logs/app.log",
			MaxSizeMB:  200,
			MaxBackups: 10,
			MaxAgeDays: 30,
			Compress:   true,
			Encoding:   lad.JSONEncoding, // default is JSONEncoding
		}),
		lad.WithCaller(),
		lad.WithStacktrace(zapcore.ErrorLevel),
	)

	defer func() { _ = lad.Sync(lad.L()) }()

	lad.S().Infow("request", "path", "/health", "ok", true)
}

func TestMustInitGlobal(t *testing.T) {
	lad.MustInitGlobal(
		lad.WithConsole(lad.ConsoleConfig{
			Level:   zap.DebugLevel,
			Colored: true,
		}),
		lad.WithFile(lad.FileConfig{
			Level:      zap.InfoLevel,
			Filename:   "./logs/app.log",
			MaxSizeMB:  200,
			MaxBackups: 10,
			MaxAgeDays: 30,
			Compress:   true,
			Encoding:   lad.JSONEncoding, // default is JSONEncoding
		}),
		lad.WithCaller(),
		lad.WithStacktrace(zapcore.ErrorLevel),
	)

	defer func() { _ = lad.Sync(lad.L()) }()

	lad.L().Info("service started")
	lad.S().Infow("request", "path", "/health", "ok", true)
}
