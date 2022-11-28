package log

import (
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	l, _ := NewLogger(WithLogPath("./Logs"), WithLogFileName("server.log"), WithMaxSize(1024*1024), WithMaxSaveTime(10*time.Second))
	//f1, _ := NewFileWriter("./Logs/Info", "server.log", 1024*1024, 0)
	//f1.SetLevels([]Level{InfoLevel})
	//f2, _ := NewFileWriter("./Logs/Error", "server.log", 1024*1024, 0)
	//f2.SetLevels([]Level{ErrorLevel})
	//f3, _ := NewFileWriter("./Logs/Warn", "server.log", 1024*1024, 0)
	//f3.SetLevels([]Level{WarnLevel})
	//l.AddWriter(f1)
	//l.AddWriter(f2)
	//l.AddWriter(f3)
	for i := 0; i < 100000; i++ {
		go func(i int) {
			l.Info("hello world print ", i)
			l.Warn("hello world print ", i)
			l.Error("hello world print ", i)
		}(i)
	}
	time.Sleep(60 * time.Second)
	l.Close()
}
