package mlog

import (
	"github.com/rs/zerolog"
	"os"
)

const (
	DebugLevel = iota
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel
	// NoLevel defines an absent log level.
	NoLevel
	// Disabled disables the logger.
	Disabled
)

func init() {
	//zerolog.TimeFieldFormat = ""
	zerolog.TimestampFieldName = "t"
	zerolog.LevelFieldName = "l"
	zerolog.MessageFieldName = "m"
}

type MLog struct {
	*MLogWriter
	zerolog.Logger
}

func NewMLog(lv int, options ...Option) *MLog {
	wr := NewLogWriter(options...)
	zlog := zerolog.New(wr).With().Timestamp().Logger()
	zlog.Level(zerolog.Level(lv))

	return &MLog{
		MLogWriter: wr,
		Logger:     zlog,
	}
}

func (ml *MLog) Sync() {
	ml.Logger.Output(os.Stderr)
	ml.MLogWriter.Sync()
}
