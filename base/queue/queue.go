package queue

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"unsafe"
)

type node struct {
	v    interface{}
	next *node
}

type Queue struct {
	head *node
	tail *node
}

func NewQueue() *Queue {
	n := &node{}
	return &Queue{
		head: n,
		tail: n,
	}
}

func (q *Queue) EnQueue(v interface{}) {
	newNode := &node{
		v: v,
	}
	prev := (*node)(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.tail)), unsafe.Pointer(newNode)))
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&prev.next)), unsafe.Pointer(newNode))
}

func (q *Queue) OutQueue() (v interface{}) {
	head := q.head
	next := (*node)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&head.next))))
	if next != nil {
		q.head = next
		v = next.v
		next.v = nil
		return
	}
	return
}

type callInfo struct {
	name string
	cb   func()
}

type EventQueue interface {
	StartLoop()
	Stop()
	Post(f func(), name string)
	SetPanicHandler(f func(interface{}))
}

type eventQueue struct {
	*Queue
	wg           sync.WaitGroup
	panicHandler func(interface{})
	exit         chan struct{}
	eventIn      int32
	wakeUpChan   chan struct{}
}

func NewEventQueue() EventQueue {
	return &eventQueue{
		Queue: NewQueue(),
		panicHandler: func(i interface{}) {
			fmt.Printf("%v  \n%s\n", i, string(debug.Stack()))
			debug.PrintStack()
		},
		exit:       make(chan struct{}),
		wakeUpChan: make(chan struct{}),
	}
}

func (q *eventQueue) StartLoop() {
	q.asyncDo(func() {
		for {
			if !q.Loop() {
				break
			}
		}
	})
}

func (q *eventQueue) Stop() {
	q.exit <- struct{}{}
	q.wg.Wait()
}

func (q *eventQueue) Post(f func(), name string) {
	if f == nil {
		return
	}
	c := &callInfo{
		name: name,
		cb:   f,
	}
	q.EnQueue(c)
	if atomic.CompareAndSwapInt32(&q.eventIn, 0, 1) {
		q.wakeUpChan <- struct{}{}
	}
}

func (q *eventQueue) asyncDo(f func()) {
	q.wg.Add(1)
	go func() {
		f()
		q.wg.Done()
	}()
}

func (q *eventQueue) Loop() bool {
	defer func() {
		if r := recover(); r != nil {
			q.panicHandler(r)
		}
	}()
	select {
	case <-q.wakeUpChan:
		q.consume()
	case <-q.exit:
		return false
	}
	return true
}

func (q *eventQueue) safeCall(callInfo *callInfo) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			err := fmt.Errorf("%v: %v", r, string(buf[:l]))
			q.panicHandler(err)
		}
	}()
	if callInfo.cb != nil {
		callInfo.cb()
	}
}

func (q *eventQueue) consume() {
	atomic.StoreInt32(&q.eventIn, 0)
	for v := q.Queue.OutQueue(); v != nil; v = q.Queue.OutQueue() {
		q.safeCall(v.(*callInfo))
	}
}

func (q *eventQueue) SetPanicHandler(f func(i interface{})) {
	q.panicHandler = f
}
