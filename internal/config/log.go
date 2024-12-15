package config

import (
	"io"
	"log/slog"
	"time"

	"gabe565.com/utils/termx"
	"github.com/lmittmann/tint"
)

//go:generate go run github.com/dmarkham/enumer -type LogFormat -trimprefix Format -transform lower -text

type LogFormat uint8

const (
	FormatAuto LogFormat = iota
	FormatColor
	FormatPlain
	FormatJSON
)

func (c *Config) InitLog(w io.Writer) {
	var level slog.Level
	if c.LogLevel == "none" {
		level = slog.LevelError + 1
	} else if err := level.UnmarshalText([]byte(c.LogLevel)); err != nil {
		slog.Warn("Invalid log level. Defaulting to info.", "value", c.LogLevel)
		level = slog.LevelInfo
		c.LogLevel = level.String()
	}

	var format LogFormat
	if err := format.UnmarshalText([]byte(c.LogFormat)); err != nil {
		slog.Warn("Invalid log format. Defaulting to auto.", "value", c.LogFormat)
		format = FormatAuto
		c.LogFormat = format.String()
	}

	InitLog(w, level, format)
}

func InitLog(w io.Writer, level slog.Level, format LogFormat) {
	switch format {
	case FormatJSON:
		slog.SetDefault(slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: level,
		})))
	default:
		var color bool
		switch format {
		case FormatAuto:
			color = termx.IsColor(w)
		case FormatColor:
			color = true
		}

		slog.SetDefault(slog.New(
			tint.NewHandler(w, &tint.Options{
				Level:      level,
				TimeFormat: time.DateTime,
				NoColor:    !color,
			}),
		))
	}
}
