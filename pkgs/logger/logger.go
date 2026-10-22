package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
)

const LevelFatal = slog.Level(12)

type Logger struct {
	*slog.Logger
}

var log = Logger{slog.New(slog.NewTextHandler(os.Stderr, nil))}

type Config struct {
	Level       slog.Level
	ServiceName string
	Pretty      bool
}

func Init(cfg Config) {
	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.Level <= slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				if lvl, ok := a.Value.Any().(slog.Level); ok && lvl == LevelFatal {
					a.Value = slog.StringValue("FATAL")
				}
			}
			if a.Key == slog.TimeKey && cfg.Pretty {
				a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339))
			}
			return a
		},
	}

	var w io.Writer = os.Stderr
	var local slog.Handler
	if cfg.Pretty {
		local = slog.NewTextHandler(w, opts)
	} else {
		local = slog.NewJSONHandler(w, opts)
	}

	otelHandler := &levelHandler{
		level:   cfg.Level,
		Handler: otelslog.NewHandler(cfg.ServiceName, otelslog.WithLoggerProvider(otel.GetLoggerProvider())),
	}

	log.Logger = slog.New(slog.NewMultiHandler(local, otelHandler))
}

func Get() *Logger { return &log }

// Fatal logs at LevelFatal and exits with status 1.
func (l *Logger) Fatal(msg string, args ...any) {
	ctx := context.Background()
	if l.Enabled(ctx, LevelFatal) {
		var pcs [1]uintptr
		runtime.Callers(2, pcs[:]) // skip runtime.Callers and Fatal
		r := slog.NewRecord(time.Now(), LevelFatal, msg, pcs[0])
		r.Add(args...)
		_ = l.Handler().Handle(ctx, r)
	}
	os.Exit(1)
}

type levelHandler struct {
	level slog.Level
	slog.Handler
}

func (h *levelHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return lvl >= h.level && h.Handler.Enabled(ctx, lvl)
}

func (h *levelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelHandler{level: h.level, Handler: h.Handler.WithAttrs(attrs)}
}

func (h *levelHandler) WithGroup(name string) slog.Handler {
	return &levelHandler{level: h.level, Handler: h.Handler.WithGroup(name)}
}
