package network

import (
	"jnet/network/base"
	"sync"
	"sync/atomic"
)

type SessionManager struct {
	sessions sync.Map  //所有链接
	Pool     sync.Pool //临时对象池
	Incr     uint64    //递增号
	sync.RWMutex
}

func (s *SessionManager) Store(id uint64, session interface{}) {
	s.sessions.Store(id, session)
}
func (s *SessionManager) GetIncrID() uint64 {
	return atomic.AddUint64(&s.Incr, 1)
}
func (s *SessionManager) Del(id uint64) {
	s.sessions.Delete(id)
}

func (s *SessionManager) ClearConn() {
	s.sessions.Range(func(key, value interface{}) bool {
		ses, ok := value.(base.Session)
		if ok {
			ses.Close()
		}
		keyUint64, ok := key.(uint64)
		if ok {
			s.Del(keyUint64)
		}
		return true
	})
}
