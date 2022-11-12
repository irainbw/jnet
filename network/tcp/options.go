package tcp

import (
	"jnet/network/base"
	"time"
)

type Option func(s *SvrOpt)

type SvrOpt struct {
	keepTcpAlive      time.Duration
	receiveBufferSize int //单次接收缓存
	Codec             base.Codec
}

func loadAllOptions(ops ...Option) *SvrOpt {
	opts := new(SvrOpt)
	for _, op := range ops {
		op(opts)
	}
	return opts

}

func WithKeepTcpAlive(time time.Duration) Option {
	return func(s *SvrOpt) {
		s.keepTcpAlive = time
	}
}

func WithReceiveBufferSize(size int) Option {
	return func(s *SvrOpt) {
		s.receiveBufferSize = size
	}
}

func WithCodec(codec base.Codec) Option {
	return func(s *SvrOpt) {
		s.Codec = codec
	}
}
