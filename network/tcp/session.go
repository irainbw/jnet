package tcp

import (
	"bytes"
	"errors"
	"fmt"
	"jnet/network/base"
	"net"
	"sync"
	"sync/atomic"
)

const (
	state_null = iota
	state_run
	state_stop
)

type session struct {
	base.SessionIdentify
	conn       net.Conn
	server     *Server
	msgChan    chan []byte
	Codec      base.Codec
	recvBuffer *bytes.Buffer
	state      int32
	property   sync.Map
}

func newSession(conn net.Conn, s *Server) *session {
	obj := s.Pool.Get()
	ses := obj.(*session)
	ses.conn = conn
	ses.Codec = s.Codec
	ses.recvBuffer = new(bytes.Buffer)
	ses.msgChan = make(chan []byte, 1024)
	ses.SetID(s.GetIncrID())
	s.Store(ses.ID(), ses)
	ses.server = s
	tc, ok := conn.(*net.TCPConn)
	if !ok {
		return ses
	}
	if s.SvrOpt.keepTcpAlive > 0 {
		_ = tc.SetKeepAlive(true)
		_ = tc.SetKeepAlivePeriod(s.SvrOpt.keepTcpAlive)
	}
	return ses
}

func (s *session) Close() {
	if atomic.CompareAndSwapInt32(&s.state, state_run, state_stop) {
		_ = s.conn.Close()
	}
}

func (s *session) SetState(state int32) {
	atomic.StoreInt32(&s.state, state)
}

func (s *session) Start() {
	s.SetState(state_run)
	s.server.HandlePacket(&base.Request{
		Ses: s,
		Msg: base.NewMsgPackage(base.SessionConnect, nil),
	})
	go s.StartReader()
	go s.StartWriter()
}

func (s *session) StartReader() {
	//单次最大接收
	var packet []byte
	if s.server.receiveBufferSize > 0 {
		packet = make([]byte, s.server.receiveBufferSize)
	} else {
		packet = make([]byte, 65536)
	}

	for {
		n, err := s.conn.Read(packet[:])
		if err != nil {
			break
		}
		s.recvBuffer.Write(packet[:n])
		err = s.processRead()
		if err != nil {
			fmt.Println("session ID read err ", err.Error())
			break
		}
	}
	close(s.msgChan)
	s.server.recycleSession(s)
	fmt.Println(s.ID(), "read close")
}

func (s *session) StartWriter() {
	defer func() {
		fmt.Println(s.ID(), "write close")
	}()
	for {
		select {
		case data, ok := <-s.msgChan:
			if ok {
				if _, err := s.conn.Write(data); err != nil {
					return
				}
			} else {
				return
			}
		}
	}

}

func (s *session) Next(n int) []byte {
	return s.recvBuffer.Next(n)
}
func (s *session) Read() []byte {
	return s.recvBuffer.Bytes()
}
func (s *session) Send(msgID uint32, data []byte) error {
	if s.state == state_stop {
		return errors.New("session closed")
	}
	rawMsg, err := s.Codec.Encode(base.NewMsgPackage(msgID, data))
	if err != nil {
		return err
	}
	select {
	case s.msgChan <- rawMsg:
	default:
		s.Close()
	}
	return nil
}

func (s *session) read() (base.IMessage, error) {
	return s.Codec.Decode(s)
}
func (s *session) processRead() (err error) {

	for {
		decodeMsg, er := s.read()
		if er != nil {
			err = er
			break
		}
		if decodeMsg == nil {
			break
		}
		//handleMsg
		req := &base.Request{
			Ses: s,
			Msg: decodeMsg,
		}
		s.server.HandlePacket(req)
	}
	return
}
