package lad

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestTrimPathFromMarker(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		marker string
		want   string
		ok     bool
	}{
		{
			name:   "match middle marker",
			path:   "/Users/me/work/omivix/application/service.go",
			marker: "omivix",
			want:   "omivix/application/service.go",
			ok:     true,
		},
		{
			name:   "marker as prefix",
			path:   "omivix/application/service.go",
			marker: "omivix",
			want:   "omivix/application/service.go",
			ok:     true,
		},
		{
			name:   "not found",
			path:   "/Users/me/work/project/application/service.go",
			marker: "omivix",
			want:   "",
			ok:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := trimPathFromMarker(normalizePathPart(tt.path), normalizePathPart(tt.marker))
			if ok != tt.ok {
				t.Fatalf("ok=%v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("got=%q, want %q", got, tt.want)
			}
		})
	}
}

func TestWithCallerPathFrom_OrderIndependent(t *testing.T) {
	t.Run("path option after file option", func(t *testing.T) {
		logFile := filepath.Join(t.TempDir(), "a.log")
		caller := writeAndReadCaller(t, logFile,
			WithFile(FileConfig{
				Level:    zapcore.InfoLevel,
				Filename: logFile,
				Encoding: JSONEncoding,
			}),
			WithCaller(),
			WithCallerPathFrom("omivix"),
		)
		if !strings.HasPrefix(caller, "omivix/") {
			t.Fatalf("caller=%q, want prefix omivix/", caller)
		}
	})

	t.Run("path option before file option", func(t *testing.T) {
		logFile := filepath.Join(t.TempDir(), "b.log")
		caller := writeAndReadCaller(t, logFile,
			WithCallerPathFrom("omivix"),
			WithFile(FileConfig{
				Level:    zapcore.InfoLevel,
				Filename: logFile,
				Encoding: JSONEncoding,
			}),
			WithCaller(),
		)
		if !strings.HasPrefix(caller, "omivix/") {
			t.Fatalf("caller=%q, want prefix omivix/", caller)
		}
	})
}

func writeAndReadCaller(t *testing.T, file string, opts ...Option) string {
	t.Helper()

	l, err := New(opts...)
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}
	defer func() { _ = Sync(l) }()

	logFromTestHelper(l)
	_ = Sync(l)

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		t.Fatal("log file is empty")
	}

	var row map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &row); err != nil {
		t.Fatalf("parse json log line: %v", err)
	}
	caller, _ := row["caller"].(string)
	if caller == "" {
		t.Fatalf("caller field missing: %v", row)
	}
	return caller
}

func logFromTestHelper(l *Logger) {
	l.Info("probe")
}
