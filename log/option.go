package log

import (
	"time"
)

type Option func(l *LogConfig)

type LogConfig struct {
	stdout              bool
	logLevel            Level         // 需要记录的日志级别
	maxSaveTime         time.Duration // 日志文件保存最长时间
	maxSize             int64         // 日志文件大小上限
	name                string        // 日志来自哪个应用
	logPath             string        // 日志存储路径
	logFileName         string        // 日志文件名
	DisableReportCaller bool
}

func (l *LogConfig) LoadAllConfig(op []Option) {
	for _, opFunc := range op {
		opFunc(l)
	}
}

func WithStdout(b bool) Option {
	return func(l *LogConfig) {
		l.stdout = b
	}
}

func WithLogLevel(level Level) Option {
	return func(l *LogConfig) {
		l.logLevel = level
	}
}

func WithMaxSaveTime(t time.Duration) Option {
	return func(l *LogConfig) {
		l.maxSaveTime = t
	}
}

func WithMaxSize(size int64) Option {
	return func(l *LogConfig) {
		l.maxSize = size
	}
}

func WithName(name string) Option {
	return func(l *LogConfig) {
		l.name = name
	}
}

func WithLogFileName(fileName string) Option {
	return func(l *LogConfig) {
		l.logFileName = fileName
	}
}

func WithLogPath(filepath string) Option {
	return func(l *LogConfig) {
		l.logPath = filepath
	}
}

func WithDisableReportCaller(b bool) Option {
	return func(l *LogConfig) {
		l.DisableReportCaller = b
	}
}
