package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/bridges/otelzerolog"
	"go.opentelemetry.io/otel/log/global"
)

var log zerolog.Logger

type Config struct {
	Level       zerolog.Level
	ServiceName string
	Pretty      bool
}

func Init(cfg Config) {
	var w io.Writer = os.Stderr
	if cfg.Pretty {
		w = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	}

	base := zerolog.New(w).Level(cfg.Level).With().Timestamp().Logger()

	hook := otelzerolog.NewHook(cfg.ServiceName, otelzerolog.WithLoggerProvider(global.GetLoggerProvider()))
	log = base.Hook(hook)
}

func Get() *zerolog.Logger { return &log }
