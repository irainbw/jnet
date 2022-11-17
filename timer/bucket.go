package timer

import (
	"container/list"
	"sync"
	"sync/atomic"
	"unsafe"
)

type Timer struct {
	expiration int64
	task       func()
	b          unsafe.Pointer //*bucket
	element    *list.Element
	canceled   int32
}

func (t *Timer) getBucket() *bucket {
	return (*bucket)(atomic.LoadPointer(&t.b))
}

func (t *Timer) setBucket(b *bucket) {
	atomic.StorePointer(&t.b, unsafe.Pointer(b))
}

func (t *Timer) Stop() {
	atomic.StoreInt32(&t.canceled, 1)
}

//相同过期时间的任务
type bucket struct {
	expiration int64
	mux        sync.Mutex
	timers     *list.List
}

func newBucket() *bucket {
	return &bucket{
		expiration: -1,
		timers:     list.New(),
	}
}

func (b *bucket) Add(t *Timer) {
	b.mux.Lock()
	e := b.timers.PushBack(t)
	t.setBucket(b)
	t.element = e
	b.mux.Unlock()
}
func (b *bucket) Expiration() int64 {
	return atomic.LoadInt64(&b.expiration)
}

/*func (b *bucket) Remove(t *Timer) bool {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.remove(t)
}*/

func (b *bucket) remove(t *Timer) bool {
	if t.getBucket() != b {
		return false
	}
	b.timers.Remove(t.element)
	t.setBucket(nil)
	t.element = nil
	return true
}

func (b *bucket) SetExpiration(expiration int64) bool {
	return atomic.SwapInt64(&b.expiration, expiration) != expiration
}

func (b *bucket) Refresh(foreach func(timer *Timer)) {
	var timerArr []*Timer
	b.mux.Lock()
	for e := b.timers.Front(); e != nil; {
		next := e.Next() // 遍历过程中移除元素
		t := e.Value.(*Timer)
		b.remove(t)
		timerArr = append(timerArr, t)
		e = next
	}
	b.mux.Unlock()
	b.SetExpiration(-1)
	for _, t := range timerArr {
		foreach(t)
	}
}
