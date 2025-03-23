package ringqueue

import (
	"fmt"
)

var ErrFullQueue = fmt.Errorf("queue is full")

type unsafeRQ[T any] struct {
	data   []T  // container data of a generic type T
	isFull bool // disambiguate whether the queue is full or empty
	start  int  // start index (inclusive, i.e. first element)
	end    int  // end index (exclusive, i.e. next after last element)

	whenFull WhenFull
}

type WhenFull int

const (
	WhenFullError = WhenFull(iota)
	WhenFullOverwrite
)

func NewUnsafe[T any](capacity int, whenFull WhenFull) (RingQueue[T], error) {
	return newUnsafe[T](capacity, whenFull)
}

func newUnsafe[T any](capacity int, whenFull WhenFull) (*unsafeRQ[T], error) {
	rq := &unsafeRQ[T]{
		data:     make([]T, capacity),
		isFull:   false,
		start:    0,
		end:      0,
		whenFull: whenFull,
	}
	return rq, nil
}

func (r *unsafeRQ[T]) String() string {
	return fmt.Sprintf(
		"[RQ full:%v size:%d start:%d end:%d data:%v]",
		r.isFull,
		len(r.data),
		r.start,
		r.end,
		r.data)
}

func (r *unsafeRQ[T]) Push(elem T) (int, error) {
	if r.isFull {
		switch r.whenFull {
		case WhenFullError:
			return 0, ErrFullQueue
		case WhenFullOverwrite:
			// continue pushing
			break
		default:
			panic("unhandled default case")
		}
	}

	r.data[r.end] = elem              // place the new element on the available space
	r.end = (r.end + 1) % len(r.data) // move the end forward by modulo of capacity
	r.isFull = r.end == r.start       // check if we're full now

	return r.Len(), nil
}

func (r *unsafeRQ[T]) Pop() (T, int, bool) {
	var res T // "zero" element (respective of the type)
	if !r.isFull && r.start == r.end {
		return res, 0, false
	}

	res = r.data[r.start]                 // copy over the first element in the queue
	r.start = (r.start + 1) % len(r.data) // move the start of the queue
	r.isFull = false                      // since we're removing elements, we can never be full

	return res, r.Len(), true
}

func (r *unsafeRQ[T]) Peek() (T, int, bool) {
	var res T // "zero" element (respective of the type)
	if !r.isFull && r.start == r.end {
		return res, 0, false
	}
	return r.data[r.start], r.Len(), true
}

func (r *unsafeRQ[T]) Len() int {
	res := r.end - r.start
	if res < 0 || (res == 0 && r.isFull) {
		res = len(r.data) - res
	}
	return res
}
func (r *unsafeRQ[T]) Cap() int {
	return len(r.data)
}

func (r *unsafeRQ[T]) IsFull() bool {
	return r.isFull
}
