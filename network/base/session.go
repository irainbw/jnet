package base

type SessionIdentify struct {
	id uint64
}

func (s *SessionIdentify) ID() uint64 {
	return s.id
}

func (s *SessionIdentify) SetID(id uint64) {
	s.id = id
}

type Session interface {
	ID() uint64
	Close()
	Next(n int) []byte
	Read() []byte
}
