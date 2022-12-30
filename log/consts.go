package log

import "sync"

type (
	Level int
)

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

var AllLevels = []Level{
	PanicLevel,
	FatalLevel,
	ErrorLevel,
	WarnLevel,
	InfoLevel,
	DebugLevel,
}

var levelString = [...]string{
	"[PANIC]",
	"[FATAL]",
	"[ERROR]",
	"[WARN]",
	"[INFO]",
	"[DEBUG]",
}

const (
	SlashFormat         = "2006/01/02 15:04:05"
	SlashWithMillFormat = "2006/01/02 15:04:05.000000"
	DashFormat          = "2006-01-02 15:04:05"
	DashWithMillFormat  = "2006-01-02 15:04:05.000000"
)

func (level Level) String() string {
	if len(levelString) < int(level) {
		return "unknown"
	}
	return levelString[level]
}
func (level Level) valid() bool {
	return level >= PanicLevel && level <= DebugLevel
}

type WgWrapper struct {
	sync.WaitGroup
}

func (w *WgWrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}
