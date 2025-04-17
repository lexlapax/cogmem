package log

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestInfoLogs(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))
	Info("hello %s", "world")
	out := buf.String()
	if !strings.Contains(out, "INFO") {
		t.Errorf("expected INFO in log output, got %q", out)
	}
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected message in log output, got %q", out)
	}
}

func TestErrorLogs(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError})
	slog.SetDefault(slog.New(handler))
	Error("error %s", "test")
	out := buf.String()
	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected ERROR in log output, got %q", out)
	}
	if !strings.Contains(out, "error test") {
		t.Errorf("expected message in log output, got %q", out)
	}
}
