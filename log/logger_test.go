package log

import "testing"

func TestEntry_Log(t *testing.T) {
	logger := New()
	logger.SetReportCaller(true)
	for i := 0; i < 10000; i++ {
		logger.Info("hello ", i)
	}
}
