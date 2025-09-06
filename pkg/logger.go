package util

import (
	"log/slog"
	"os"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

type slogWrapper struct {
	logger *slog.Logger
}

func (s *slogWrapper) Debug(msg string, args ...any) {
	s.logger.Debug(msg, args...)
}

func (s *slogWrapper) Info(msg string, args ...any) {
	s.logger.Info(msg, args...)
}

func (s *slogWrapper) Warn(msg string, args ...any) {
	s.logger.Warn(msg, args...)
}

func (s *slogWrapper) Error(msg string, args ...any) {
	s.logger.Error(msg, args...)
}

func (s *slogWrapper) With(args ...any) Logger {
	return &slogWrapper{logger: s.logger.With(args...)}
}

func NewLogger(component string) Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})).With("component", component)

	return &slogWrapper{logger: logger}
}

func NewDevLogger(component string) Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})).With("component", component)

	return &slogWrapper{logger: logger}
}
