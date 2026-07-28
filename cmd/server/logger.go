package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

type coloredHandler struct {
	w io.Writer
}

func (h coloredHandler) Enabled(ctx context.Context, l slog.Level) bool { return true }

func (h coloredHandler) Handle(ctx context.Context, r slog.Record) error {
	colors := map[slog.Level]string{
		slog.LevelDebug: "\x1b[32m",   // green
		slog.LevelInfo:  "\x1b[34m",   // blue
		slog.LevelWarn:  "\x1b[0;93m", // yellow
		slog.LevelError: "\x1b[31m",   // red
	}
	reset := "\x1b[0m"
	keyColor := "\x1b[36m"

	levelStr := map[slog.Level]string{
		slog.LevelDebug: "DBG",
		slog.LevelInfo:  "INFO",
		slog.LevelWarn:  "WARN",
		slog.LevelError: "ERR",
	}[r.Level]

	ts := r.Time.Format("06.01.02 T15:04:05")

	msg := r.Message
	r.Attrs(func(a slog.Attr) bool {
		msg += fmt.Sprintf(" %s%s%s=%v", keyColor, a.Key, reset, a.Value.Any())
		return true
	})

	_, err := fmt.Fprintf(h.w, "%s [%s%s%s] %s\n", ts, colors[r.Level], levelStr, reset, msg)
	return err
}

func (h coloredHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h coloredHandler) WithGroup(name string) slog.Handler       { return h }
