package tcp

import (
	"encoding/binary"
	"fmt"
	"jnet/base/vector"
	"jnet/network"
	"jnet/network/base"
	"sync"
	"time"
)

type PacketFunc func(request base.IRequest) bool //回调函数
const Internal = 10 * time.Second

type Server struct {
	l *listener
	*SvrOpt
	network.SessionManager
	protoAddr      string
	packetFuncList *vector.Vector
}

func NewServer(protoAddr string, opt ...Option) *Server {
	svr := new(Server)
	svr.SvrOpt = loadAllOptions(opt...)
	svr.protoAddr = protoAddr
	svr.packetFuncList = vector.NewVector()
	svr.SessionManager = network.SessionManager{
		Pool: sync.Pool{
			New: func() interface{} {
				return &session{}
			}},
	}
	if svr.Codec == nil {
		svr.Codec = &base.PacketParser{
			PacketHeadLen: 8, //uint32+uint32
			MaxPacketLen:  40960,
			ByteOrder:     binary.BigEndian,
		}
	}
	return svr
}

func (s *Server) Serve() {
	go func() {
		for {
			err := s.startListen()
			if err != nil {
				return
			}
			s.startAccept()
			time.Sleep(Internal)
		}
	}()
}

func (s *Server) startListen() error {
	l, err := newListener(s.protoAddr)
	if err != nil {
		return err
	}
	s.l = l
	return nil
}

func (s *Server) BindPacketFunc(callfunc PacketFunc) {
	s.packetFuncList.PushBack(callfunc)
}

func (s *Server) HandlePacket(req base.IRequest) {
	for _, v := range s.packetFuncList.Values() {
		if v.(PacketFunc)(req) {
			break
		}
	}
}

func (s *Server) startAccept() {
	for {
		conn, err := s.l.ln.AcceptTCP()
		if err != nil {
			fmt.Println("Accept err ", err)
			break
		}
		ses := newSession(conn, s)
		ses.Start()
	}
}

func (s *Server) Close() {
	_ = s.l.ln.Close()
	s.SessionManager.ClearConn()
}

func (s *Server) recycleSession(session *session) {
	s.HandlePacket(&base.Request{
		Ses: session,
		Msg: base.NewMsgPackage(base.SessionClose, nil),
	})
	session.Close()
	s.Del(session.ID())
	s.Pool.Put(session)
}
