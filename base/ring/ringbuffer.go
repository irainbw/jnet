package ring

import (
	"errors"
)

type ByteBuffer struct {
	data     []byte
	size     int
	writePos int
	readPos  int
	empty    bool
}

func NewRingBuffer(size int) *ByteBuffer {
	return &ByteBuffer{
		data:  make([]byte, size),
		size:  size,
		empty: true,
	}
}
func (r *ByteBuffer) AvailableReadLen() int {
	if r.readPos == r.writePos {
		if r.empty {
			return 0
		}
		return r.size
	}
	if r.readPos < r.writePos {
		return r.writePos - r.readPos
	}
	return r.size - r.readPos + r.writePos
}
func (r *ByteBuffer) AvailableWriteLen() int {
	if r.writePos == r.readPos {
		if r.empty {
			return r.size
		}
		return 0
	}
	if r.writePos < r.readPos {
		return r.readPos - r.writePos
	}
	return r.size - r.writePos + r.readPos
}

func (r *ByteBuffer) LazyRead(n int) (head []byte, tail []byte) {
	if r.empty {
		return
	}
	if n <= 0 {
		return
	}
	readLen := r.AvailableReadLen()
	if readLen > n {
		readLen = n
	}
	if r.readPos+readLen <= r.size {
		head = r.data[r.readPos : r.readPos+readLen]
		return
	}
	headSize := r.size - r.readPos
	head = r.data[r.readPos:]
	tailSize := readLen - headSize
	tail = r.data[:tailSize]
	return
}
func (r *ByteBuffer) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	if r.empty {
		return 0, errors.New("empty ringbuffer")
	}
	n = r.AvailableReadLen()
	if n > len(b) {
		n = len(b)
	}
	if r.readPos+n <= r.size {
		copy(b, r.data[r.readPos:r.readPos+n])
	} else {
		headSize := r.size - r.readPos
		copy(b, r.data[r.readPos:])
		tailSize := n - headSize
		copy(b[headSize:], r.data[:tailSize])
	}
	r.readPos = (r.readPos + n) % r.size
	if r.readPos == r.writePos {
		r.Reset()
	}
	return
}
func (r *ByteBuffer) Write(b []byte) (n int, err error) {
	n = len(b)
	if n == 0 {
		return
	}
	writeLen := r.AvailableWriteLen()
	if n > writeLen {
		r.resize(n - writeLen)
	}
	if r.writePos+n <= r.size {
		copy(r.data[r.writePos:], b)
	} else {
		headSize := r.size - r.readPos
		copy(r.data[r.writePos:], b[:headSize])
		copy(r.data, b[headSize:])
	}
	r.writePos = (r.writePos + n) % r.size
	r.empty = false
	return
}
func (r *ByteBuffer) Reset() {
	newCap := r.size >> 1
	r.data = make([]byte, newCap)
	r.size = newCap
	r.empty = true
	r.readPos = 0
	r.writePos = 0
}
func (r *ByteBuffer) resize(size int) {
	newSize := r.size + size
	newData := make([]byte, newSize)
	n, _ := r.Read(newData)
	r.data = newData
	r.size = newSize
	r.readPos = 0
	r.writePos = n
}
func (r *ByteBuffer) Shift(n int) {
	if n <= 0 {
		return
	}
	if n < r.AvailableReadLen() {
		r.readPos = (r.readPos + n) % r.size
	} else {
		r.Reset()
	}
}
func (r *ByteBuffer) Empty() bool {
	return r.empty
}
