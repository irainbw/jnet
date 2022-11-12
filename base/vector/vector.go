package vector

const (
	VectorBlockSize = 16
)

type Vector struct {
	arr   []interface{}
	size  int
	count int
}

func NewVector() *Vector {
	return &Vector{}
}

func (v *Vector) insert(index int) {
	if v.size == v.count {
		v.resize(v.count + 1)
	} else {
		v.count++
	}
	for i := v.count - 1; i > index; i-- {
		v.arr[i] = v.arr[i-1]
	}
}

func (v *Vector) resize(newCount int) {
	if newCount < 0 {
		return
	}
	blocks := newCount / VectorBlockSize
	if newCount%VectorBlockSize != 0 {
		blocks++
	}
	v.count = newCount
	v.size = blocks * VectorBlockSize
	newArray := make([]interface{}, v.size+1)
	copy(newArray, v.arr)
	v.arr = newArray
}

func (v *Vector) PushFront(value interface{}) {
	v.insert(0)
	v.arr[0] = value
}
func (v *Vector) PushBack(value interface{}) {
	if v.count == v.size {
		v.resize(v.count + 1)
	} else {
		v.count++
	}
	v.arr[v.count-1] = value
}

func (v *Vector) Values() []interface{} {
	return v.arr[0:v.count]
}
