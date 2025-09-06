package util

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger("test-component")
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}

	wrapper, ok := logger.(*slogWrapper)
	if !ok {
		t.Fatal("NewLogger did not return slogWrapper")
	}
	if wrapper.logger == nil {
		t.Fatal("slogWrapper logger is nil")
	}
}

func TestNewDevLogger(t *testing.T) {
	logger := NewDevLogger("dev-component")
	if logger == nil {
		t.Fatal("NewDevLogger returned nil")
	}

	wrapper, ok := logger.(*slogWrapper)
	if !ok {
		t.Fatal("NewDevLogger did not return slogWrapper")
	}
	if wrapper.logger == nil {
		t.Fatal("slogWrapper logger is nil")
	}
}

func TestSlogWrapper_LogMethods(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &slogWrapper{
		logger: slog.New(handler),
	}

	tests := []struct {
		name     string
		logFunc  func()
		expected string
	}{
		{
			name: "debug log",
			logFunc: func() {
				logger.Debug("debug message", "key", "value")
			},
			expected: "debug message",
		},
		{
			name: "info log",
			logFunc: func() {
				logger.Info("info message", "key", "value")
			},
			expected: "info message",
		},
		{
			name: "warn log",
			logFunc: func() {
				logger.Warn("warn message", "key", "value")
			},
			expected: "warn message",
		},
		{
			name: "error log",
			logFunc: func() {
				logger.Error("error message", "key", "value")
			},
			expected: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			output := buf.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Log output does not contain expected message: %q, got: %q", tt.expected, output)
			}

			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
				t.Errorf("Log output is not valid JSON: %v", err)
			}

			if logEntry["msg"] != tt.expected {
				t.Errorf("Log message = %q, want %q", logEntry["msg"], tt.expected)
			}
			if logEntry["key"] != "value" {
				t.Errorf("Log key = %q, want %q", logEntry["key"], "value")
			}
		})
	}
}

func TestSlogWrapper_With(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &slogWrapper{
		logger: slog.New(handler),
	}

	childLogger := logger.With("component", "test-component", "id", 123)
	if childLogger == nil {
		t.Fatal("With returned nil")
	}

	childWrapper, ok := childLogger.(*slogWrapper)
	if !ok {
		t.Fatal("With did not return slogWrapper")
	}
	if childWrapper.logger == nil {
		t.Fatal("child slogWrapper logger is nil")
	}

	childLogger.Info("test message")
	output := buf.String()

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Log output is not valid JSON: %v", err)
	}

	if logEntry["component"] != "test-component" {
		t.Errorf("Log component = %q, want %q", logEntry["component"], "test-component")
	}
	if logEntry["id"] != float64(123) {
		t.Errorf("Log id = %v, want %v", logEntry["id"], 123)
	}
}
