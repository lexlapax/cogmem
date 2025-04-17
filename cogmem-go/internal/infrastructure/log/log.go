package log

import (
	"fmt"
	"log/slog"
	"os"
)

func init() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// Info logs an informational message (supports fmt-style formatting).
func Info(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	slog.Info(msg)
}

// Error logs an error message (supports fmt-style formatting).
func Error(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	slog.Error(msg)
}

// Fatal logs a fatal error message and exits.
func Fatal(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	slog.Error(msg)
	os.Exit(1)
}
