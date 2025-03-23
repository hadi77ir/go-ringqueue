package ringqueue

import "fmt"

type RingQueue[T any] interface {
	fmt.Stringer
	Len() int
	Cap() int
	Push(element T) (newLen int, err error)
	Pop() (elem T, newLen int, ok bool)
	Peek() (elem T, len int, ok bool)
}
