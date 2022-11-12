package tcp

import (
	"errors"
	"net"
	"strings"
)

type listener struct {
	ln   *net.TCPListener
	net  string
	addr string
}

func newListener(protoAddr string) (*listener, error) {
	l := new(listener)
	l.parseProtoAddr(protoAddr)
	err := l.initial()
	return l, err
}

func (l *listener) parseProtoAddr(protoAddr string) {
	n := "tcp4"
	addr := protoAddr
	if strings.Contains(protoAddr, "://") {
		netAddr := strings.Split(protoAddr, "://")
		n = netAddr[0]
		addr = netAddr[1]
	}
	l.net = n
	l.addr = addr
}

func (l *listener) initial() error {
	var err error
	switch l.net {
	case "tcp", "tcp4", "tcp6":
		tcpAddr, err := net.ResolveTCPAddr(l.net, l.addr)
		l.ln, err = net.ListenTCP(l.net, tcpAddr)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid protoAddr")
	}
	return err
}
