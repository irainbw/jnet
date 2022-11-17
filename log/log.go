package log

import "errors"

type Level uint32
type Fields map[string]interface{}

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

var AllLevels = []Level{
	PanicLevel,
	FatalLevel,
	ErrorLevel,
	WarnLevel,
	InfoLevel,
	DebugLevel,
	TraceLevel,
}

var levelString = [...]string{
	"[PANIC]",
	"[FATAL]",
	"[ERROR]",
	"[WARN]",
	"[INFO]",
	"[DEBUG]",
	"[TRACE]",
}

func (level Level) String() string {
	if b, err := level.MarshalText(); err == nil {
		return b
	} else {
		return "unknown"
	}
}
func (level Level) MarshalText() (string, error) {
	if int(level) > len(levelString) {
		return "", errors.New("invalid level")
	}
	return levelString[level], nil
}
