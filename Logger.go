package CharmSlog

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strconv"

	clog "github.com/charmbracelet/log"
)

type CharmHandler struct {
	logger   *clog.Logger
	minLevel slog.Level
	group    string
}

func NewCharmHandler(minLevel slog.Level, prefix string) *CharmHandler {
	l := clog.NewWithOptions(os.Stderr, clog.Options{
		ReportCaller: false,
		Prefix:       prefix,
	})

	return &CharmHandler{
		logger:   l,
		minLevel: minLevel,
	}
}

func (h *CharmHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *CharmHandler) Handle(_ context.Context, r slog.Record) error {
	if !h.Enabled(nil, r.Level) {
		return nil
	}

	fields := make([]any, 0, r.NumAttrs()+3)

	r.Attrs(func(a slog.Attr) bool {
		key := a.Key
		if h.group != "" {
			key = h.group + "." + key
		}
		fields = append(fields, key, a.Value.Any())
		return true
	})

	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		if f, ok := fs.Next(); ok {
			fields = append(fields, "source", f.File+":"+strconv.Itoa(f.Line))
		}
	}

	switch r.Level {
	case slog.LevelDebug:
		h.logger.Debug(r.Message, fields...)
	case slog.LevelInfo:
		h.logger.Info(r.Message, fields...)
	case slog.LevelWarn:
		h.logger.Warn(r.Message, fields...)
	case slog.LevelError:
		h.logger.Error(r.Message, fields...)
	default:
		h.logger.Print(r.Message, fields...)
	}
	return nil
}

func (h *CharmHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fields := make([]any, 0, len(attrs))
	for _, a := range attrs {
		key := a.Key
		if h.group != "" {
			key = h.group + "." + key
		}
		fields = append(fields, key, a.Value.Any())
	}
	newLogger := h.logger.With(fields...)
	return &CharmHandler{logger: newLogger, minLevel: h.minLevel, group: h.group}
}

// WithGroup возвращает новый Handler с префиксом для ключей.
func (h *CharmHandler) WithGroup(name string) slog.Handler {
	group := name
	if h.group != "" {
		group = h.group + "." + name
	}
	return &CharmHandler{logger: h.logger, minLevel: h.minLevel, group: group}
}
