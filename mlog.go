package mlog

import (
	"os"

	"github.com/rs/zerolog"
)

const (
	// DebugLevel defines info log level.
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
	zerolog.TimeFieldFormat = "060102T15:04:05"
	zerolog.TimestampFieldName = "t"
	zerolog.LevelFieldName = "l"
	zerolog.MessageFieldName = "m"
}

// Logger global default log, stderr
var Logger, _ = New(DebugLevel)

// MLog wrapper for zerolog
type MLog struct {
	*writer
	zerolog.Logger
}

// New log
func New(lv int, options ...Option) (*MLog, error) {
	wr, err := newWriter(options...)
	if err != nil {
		return nil, err
	}

	zLog := zerolog.New(wr).With().Timestamp().Logger()
	zLog.Level(zerolog.Level(lv))

	ml := &MLog{
		writer: wr,
		Logger: zLog,
	}

	return ml, nil
}

// Sync output bytes
func (ml *MLog) Close() {
	ml.Logger.Output(os.Stderr)
	ml.writer.Close()
}
