package vector

import (
	"fmt"
	"testing"
)

func TestNewVector(t *testing.T) {
	v := NewVector()
	for i := 0; i < 10000; i++ {
		v.PushBack(i)
	}
	fmt.Println(v.Values())

}
