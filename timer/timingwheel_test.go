package timer

import (
	"fmt"
	//"github.com/RussellLuo/timingwheel"
	"testing"
	"time"
)

func BenchmarkNewTimingWheel(b *testing.B) {
	tw, _ := NewTimingWheel(time.Millisecond, 100)
	tw.Start()
	defer tw.Stop()
	fmt.Printf("NOW timer  fires %v\n", time.Now())
	var timerArray []*Timer
	for i := 0; i < 1000000; i++ {
		t := tw.ScheduleFunc(&EveryScheduler{10 * time.Millisecond}, func() {
			fmt.Printf("The timer 10 MilliSecond fires %v\n", time.Now())
		})
		timerArray = append(timerArray, t)
	}
	time.Sleep(600 * time.Second)
}
