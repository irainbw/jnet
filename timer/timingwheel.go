package timer

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type TimingWheel struct {
	tick          int64 //时间跨度
	wheelSize     int64
	interval      int64
	currentTime   int64     //时间跨度整数倍
	buckets       []*bucket //时间格
	delayQueue    *DelayQueue
	exitC         chan struct{}
	waitGroup     sync.WaitGroup
	overflowWheel unsafe.Pointer //*TimingWheel 指向更高层的时间轮
}

func NewTimingWheel(tick time.Duration, wheelSize int64) (*TimingWheel, error) {
	tickMs := int64(tick / time.Millisecond)
	if tick < 0 {
		return nil, errors.New("tick must greater than 1ms")
	}
	startMs := timeToMs(time.Now().UTC())
	return newTimingWheel(tickMs, startMs, wheelSize, NewDelayQueue(int(wheelSize))), nil
}

func newTimingWheel(tickMs int64, startMs int64, wheelSize int64, dq *DelayQueue) *TimingWheel {
	buckets := make([]*bucket, wheelSize)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	return &TimingWheel{
		tick:        tickMs,
		wheelSize:   wheelSize,
		interval:    tickMs * wheelSize,
		currentTime: truncate(startMs, tickMs), //修剪为tickMs 的整数倍
		buckets:     buckets,
		delayQueue:  dq,
		exitC:       make(chan struct{}),
	}
}

func (tw *TimingWheel) Start() {
	tw.asyncRun(func() {
		tw.delayQueue.Poll()
	})
	tw.asyncRun(func() {
		tw.consume()
	})
}

func (tw *TimingWheel) AfterFunc(d time.Duration, f func()) *Timer {
	t := &Timer{
		expiration: timeToMs(time.Now().UTC().Add(d)),
		task:       f,
	}
	tw.addOrRun(t)
	return t
}

func (tw *TimingWheel) consume() {
	for {
		select {
		case elem := <-tw.delayQueue.C:
			e := elem.(*bucket)
			tw.advanceTime(e.Expiration())
			e.Refresh(tw.addOrRun)
		case <-tw.exitC:
			tw.delayQueue.Exit()
			return
		}
	}
}

//推进时间轮当前时间,当有过期任务时推进时间轮
func (tw *TimingWheel) advanceTime(expiration int64) {
	curTime := atomic.LoadInt64(&tw.currentTime)
	if expiration >= curTime+tw.tick {
		curTime = truncate(expiration, tw.tick)
		atomic.StoreInt64(&tw.currentTime, curTime)
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		//有上层时间轮
		if overflowWheel != nil {
			(*TimingWheel)(overflowWheel).advanceTime(curTime)
		}
	}
}

func (tw *TimingWheel) addOrRun(t *Timer) {
	if !tw.add(t) {
		defer func() {
			err := recover()
			if err != nil {
				fmt.Fprintf(os.Stderr, "timer task %v paniced : %v\n", t, err)
				debug.PrintStack()
			}
		}()
		t.task() //回调至工作线程
	}
}

func (tw *TimingWheel) add(t *Timer) bool {
	if atomic.LoadInt32(&t.canceled) == 1 {
		return true
	}
	currentTime := atomic.LoadInt64(&tw.currentTime)
	if t.expiration < currentTime+tw.tick {
		return false
	} else if t.expiration < currentTime+tw.interval {
		numTick := t.expiration / tw.tick
		bucketIndex := numTick % tw.wheelSize //找到对应时间格
		b := tw.buckets[bucketIndex]
		b.Add(t)
		if b.SetExpiration(numTick * tw.tick) { //防止重复添加
			tw.delayQueue.Offer(b, b.Expiration())
		}
		return true
	} else {
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel == nil { //下一层时间跨度为上一层时间跨度的总长interval
			atomic.CompareAndSwapPointer(&tw.overflowWheel, nil,
				unsafe.Pointer(newTimingWheel(tw.interval, currentTime, tw.wheelSize, tw.delayQueue)))
			overflowWheel = atomic.LoadPointer(&tw.overflowWheel)
		}
		return (*TimingWheel)(overflowWheel).add(t)
	}
}

func (tw *TimingWheel) asyncRun(cb func()) {
	tw.waitGroup.Add(1)
	go func() {
		cb()
		tw.waitGroup.Done()
	}()
}

func (tw *TimingWheel) Stop() {
	close(tw.exitC)
	tw.waitGroup.Wait()
}

func truncate(x, m int64) int64 {
	if m <= 0 {
		return x
	}
	return x - x%m
}

type Scheduler interface {
	Next(time.Time) time.Time
}

func (tw *TimingWheel) ScheduleFunc(s Scheduler, f func()) (t *Timer) {
	expiration := s.Next(time.Now().UTC())
	if expiration.IsZero() {
		return
	}
	t = &Timer{
		expiration: timeToMs(expiration),
		task: func() {
			expiration = s.Next(msToTime(t.expiration))
			if !expiration.IsZero() {
				t.expiration = timeToMs(expiration)
				tw.addOrRun(t)
			}
			f()
		},
	}
	tw.addOrRun(t)
	return
}

func timeToMs(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func msToTime(t int64) time.Time {
	return time.Unix(0, t*int64(time.Millisecond)).UTC()
}

type EveryScheduler struct {
	Interval time.Duration
}

func (s *EveryScheduler) Next(prev time.Time) time.Time {
	return prev.Add(s.Interval)
}
