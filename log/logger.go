package log

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Logger struct {
	config  *LogConfig
	Writers LevelWriters
	closed  int32
	mux     sync.Mutex
}

func NewLogger(opts ...Option) (*Logger, error) {
	opt := &LogConfig{
		logLevel: InfoLevel,
	}
	opt.LoadAllConfig(opts)
	logger := &Logger{
		config:  opt,
		Writers: make(LevelWriters),
	}
	if opt.stdout {
		logger.AddWriter(NewStdWriter())
	}
	logFileName := "temp.log"
	if opt.logFileName != "" {
		logFileName = opt.logFileName
	}
	if opt.logPath != "" {
		f, err := NewFileWriter(opt.logPath, logFileName, opt.maxSize, opt.maxSaveTime)
		if err != nil {
			return nil, err
		}
		logger.AddWriter(f)
	}

	return logger, nil
}

func (l *Logger) AddWriter(logWriter LogWriter) {
	l.Writers.Add(logWriter)
}

func (l *Logger) Close() {
	if atomic.LoadInt32(&l.closed) == 1 {
		return
	}
	atomic.StoreInt32(&l.closed, 1)
	for _, writers := range l.Writers {
		for _, w := range writers {
			w.Close()
		}
	}
	l.Writers = nil
}

func Copy(src []byte) (b []byte) {
	if len(src) > 0 {
		b = make([]byte, len(src))
		copy(b, src)
	}
	return
}
func (l *Logger) log(level Level, message string) {
	if !l.canOutput(level) {
		return
	}
	var (
		now         = time.Now()
		appName     = l.config.name
		levelPrefix = level.String()
		timeDesc    = now.Format(SlashWithMillFormat)
	)

	b := bufferPool.Get()
	if appName != "" {
		b.WriteString("[")
		b.WriteString(appName)
		b.WriteString("] ")
	}
	b.WriteString(levelPrefix)
	b.WriteString("[")
	b.WriteString(timeDesc)
	b.WriteString("] ")
	b.WriteString(message)
	b.WriteString("\n")
	logStr := Copy(b.Bytes())
	bufferPool.Put(b)
	l.fireWrite(level, logStr)
}

func (l *Logger) fireWrite(level Level, b []byte) {
	var tmpLevelWriters LevelWriters
	l.mux.Lock()
	tmpLevelWriters = make(LevelWriters, len(l.Writers))
	for k, v := range l.Writers {
		tmpLevelWriters[k] = v
	}
	l.mux.Unlock()
	err := tmpLevelWriters.Fire(level, b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fire hook: %v\n", err)
	}
}

func (l *Logger) exit() {
	l.Close()
	os.Exit(1)
}

func (l *Logger) canOutput(level Level) bool {
	if atomic.LoadInt32(&l.closed) == 1 {
		return false
	}
	if !level.valid() {
		return false
	}
	if l.config.logLevel < level {
		return false
	}
	return true
}

func (l *Logger) Panic(args ...interface{}) {
	l.log(PanicLevel, fmt.Sprint(args...))
}

func (l *Logger) Fatal(args ...interface{}) {
	l.log(FatalLevel, fmt.Sprint(args...))
	l.exit()
}

func (l *Logger) Debug(args ...interface{}) {
	l.log(DebugLevel, fmt.Sprint(args...))
}

func (l *Logger) Info(args ...interface{}) {
	l.log(InfoLevel, fmt.Sprint(args...))
}

func (l *Logger) Error(args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprint(args...))
}

func (l *Logger) Print(args ...interface{}) {
	l.log(InfoLevel, fmt.Sprint(args...))
}

func (l *Logger) Warn(args ...interface{}) {
	l.log(WarnLevel, fmt.Sprint(args...))
}

func (l *Logger) Panicln(args ...interface{}) {
	l.log(PanicLevel, fmt.Sprintln(args...))
}

func (l *Logger) FatalLn(args ...interface{}) {
	l.log(FatalLevel, fmt.Sprintln(args...))
}

func (l *Logger) Debugln(args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintln(args...))
}

func (l *Logger) Infoln(args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintln(args...))
}

func (l *Logger) Errorln(args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintln(args...))
}

func (l *Logger) Println(args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintln(args...))
}

func (l *Logger) Warnln(args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintln(args...))
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	l.log(PanicLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(FatalLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(format, args...))
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintf(format, args...))
}
