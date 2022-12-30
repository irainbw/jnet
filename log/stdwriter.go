package log

import (
	"os"
	"sync"
	"sync/atomic"
)

type StdWriter struct {
	writeBuffer chan []byte
	WgWrapper
	sync.Once
	closed int32
	exit   chan struct{}
}

func NewStdWriter() *StdWriter {
	s := &StdWriter{
		writeBuffer: make(chan []byte, 1<<10),
		exit:        make(chan struct{}),
	}
	s.Start()
	return s
}

func (s *StdWriter) Start() {
	s.Wrap(func() {
		for {
			select {
			case <-s.exit:
				return
			case content := <-s.writeBuffer:
				_, _ = os.Stdout.Write(content)
			}
		}

	})
}

func (s *StdWriter) Close() {
	s.Once.Do(func() {
		atomic.StoreInt32(&s.closed, 1)
		s.exit <- struct{}{}
		s.Wait()
	})
}

func (s *StdWriter) Levels() []Level {
	return AllLevels
}

func (s *StdWriter) LogWrite(b []byte) error {
	//if isClose(s.writeBuffer) {
	//	return nil
	//}
	if atomic.LoadInt32(&s.closed) == 1 {
		return nil
	}
	s.writeBuffer <- b
	return nil
}
