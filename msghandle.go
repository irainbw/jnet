package main

import (
	"fmt"
	"jnet/network/base"
	"runtime/debug"
	"strconv"
	"sync"
)

type Handler func(msg base.IRequest)
type worker struct {
	taskQueue chan func()
	exitChan  chan struct{}
}

func NewWorker(taskChanNum int32) *worker {
	return &worker{
		taskQueue: make(chan func(), taskChanNum),
		exitChan:  make(chan struct{}),
	}
}

func (w *worker) startWork() {
	go func() {
		for {
			select {
			case task := <-w.taskQueue:
				w.safeCall(task)
			case <-w.exitChan:
				return
			}
		}
	}()
}

func (w *worker) curMsgChenLen() int {
	return len(w.taskQueue)
}

func (w *worker) safeCall(f func()) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("err stack :", err)
			debug.PrintStack()
		}
	}()
	f()
}

func (w *worker) stopWork() {
	w.exitChan <- struct{}{}
}

func (w *worker) PostTask(f func()) {
	w.taskQueue <- f
}

type MsgHandle struct {
	mutex       sync.Mutex
	size        uint64
	taskChanNum int32
	mHandlers   map[uint32]Handler
	worker      []*worker
}

func NewMsgHandle(size uint64, taskChanNum int32) *MsgHandle {
	return &MsgHandle{
		size:        size,
		taskChanNum: taskChanNum,
		mHandlers:   map[uint32]Handler{},
		worker:      make([]*worker, size),
	}
}

func (mh *MsgHandle) DeliverMsg(request base.IRequest) {
	workerID := request.GetConnection().ID() % mh.size
	handler, ok := mh.mHandlers[request.GetMsgID()]
	if !ok {
		return
	}
	mh.worker[workerID].PostTask(func() {
		handler(request)
	})
}

func (mh *MsgHandle) AddHandler(msgID uint32, handler Handler) {
	if _, ok := mh.mHandlers[msgID]; ok {
		fmt.Println("repeated api , msgID = " + strconv.Itoa(int(msgID)))
	}
	mh.mHandlers[msgID] = handler
}

func (mh *MsgHandle) Start() {
	for i := 0; i < int(mh.size); i++ {
		wk := NewWorker(mh.taskChanNum)
		mh.worker[i] = wk
		wk.startWork()
	}
}

func (mh *MsgHandle) Stop() {
	for i := 0; i < int(mh.size); i++ {
		mh.worker[i].stopWork()
	}
}
