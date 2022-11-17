package timer

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"
)

type item struct {
	Value    interface{}
	Priority int64
	Index    int
}

type PriorityQueue []*item

func NewPriorityQueue(cap int) PriorityQueue {
	return make(PriorityQueue, 0, cap)
}

func (p PriorityQueue) Len() int {
	return len(p)
}

func (p PriorityQueue) Less(i, j int) bool {
	return p[i].Priority < p[j].Priority
}

func (p PriorityQueue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	p[i].Index = i
	p[j].Index = j
}

func (p *PriorityQueue) Push(x interface{}) {
	l := len(*p)
	c := cap(*p)
	if l+1 > c {
		newP := make(PriorityQueue, l, 2*c)
		copy(newP, *p)
		*p = newP
	}
	*p = (*p)[:l+1]
	v := x.(*item)
	v.Index = l
	(*p)[l] = v
}

func (p *PriorityQueue) Pop() interface{} {
	l := len(*p)
	c := cap(*p)
	if l < c/2 && c > 25 {
		newP := make(PriorityQueue, l, c/2)
		copy(newP, *p)
		*p = newP
	}
	v := (*p)[l-1]
	v.Index = -1
	*p = (*p)[:l-1]
	return v
}

func (p *PriorityQueue) PeekAndShift(value int64) (*item, int64) {
	if p.Len() == 0 {
		return nil, 0
	}
	item := (*p)[0]
	if item.Priority > value {
		return nil, item.Priority - value
	}
	heap.Remove(p, 0)

	return item, 0
}

type DelayQueue struct {
	pq         PriorityQueue
	wakeupChan chan struct{}
	mux        sync.Mutex
	C          chan interface{}
	exitChan   chan struct{}
	sleeping   int32
}

func NewDelayQueue(size int) *DelayQueue {
	return &DelayQueue{
		pq:         NewPriorityQueue(size),
		wakeupChan: make(chan struct{}),
		C:          make(chan interface{}),
		exitChan:   make(chan struct{}),
	}
}
func (d *DelayQueue) Offer(elem interface{}, expiration int64) {
	item := &item{Value: elem, Priority: expiration}
	d.mux.Lock()
	heap.Push(&d.pq, item)
	index := item.Index
	d.mux.Unlock()
	if index == 0 { // 插入的任务是最早执行的任务
		if atomic.CompareAndSwapInt32(&d.sleeping, 1, 0) { //若正处于休眠 唤醒
			d.wakeupChan <- struct{}{}
		}
	}
}

func (d *DelayQueue) Poll() {
	for {
		now := timeToMs(time.Now().UTC())
		d.mux.Lock()
		item, recentComing := d.pq.PeekAndShift(now)
		if item == nil {
			//没有任务消耗 等待任务插入重新唤醒取出任务
			atomic.StoreInt32(&d.sleeping, 1)
		}
		d.mux.Unlock()
		if item == nil {
			if recentComing == 0 { // 没有任务存在 等待任务插入唤醒
				select {
				case <-d.wakeupChan:
					continue
				case <-d.exitChan:
					goto exit
				}
			} else if recentComing > 0 { // 最近至少一个任务存在
				select {
				case <-d.wakeupChan: // 添加了一个比当前最近任务更早的任务
					continue
				case <-time.After(time.Duration(recentComing) * (time.Millisecond)): // 当前任务是最早过期的任务
					//不再需要从唤醒管道中获取 重置休眠状态
					if atomic.SwapInt32(&d.sleeping, 0) == 0 {
						//解除阻塞
						<-d.wakeupChan
					}
					continue
				case <-d.exitChan:
					goto exit
				}
			}
		}
		select {
		case d.C <- item.Value:
		case <-d.exitChan:
			goto exit
		}
	}
exit:
	atomic.StoreInt32(&d.sleeping, 0)
}
func (d *DelayQueue) Len() int64 {
	return int64(d.pq.Len())
}
func (d *DelayQueue) Exit() {
	close(d.exitChan)
}
