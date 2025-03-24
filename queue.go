package ringqueue

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type unsafeRQ[T any] struct {
	data   []T  // container data of a generic type T
	isFull bool // disambiguate whether the queue is full or empty
	start  int  // start index (inclusive, i.e. first element)
	end    int  // end index (exclusive, i.e. next after last element)

	whenFull WhenFull

	closeOnce sync.Once
	closed    bool
	onClose   OnCloseFunc[T]
}

func NewUnsafe[T any](capacity int, whenFull WhenFull, whenEmpty WhenEmpty, onCloseFunc OnCloseFunc[T]) (RingQueue[T], error) {
	if whenEmpty != WhenEmptyError {
		return nil, ErrUnsupported
	}
	return newUnsafe[T](capacity, whenFull, onCloseFunc)
}

func newUnsafe[T any](capacity int, whenFull WhenFull, onCloseFunc OnCloseFunc[T]) (*unsafeRQ[T], error) {
	rq := &unsafeRQ[T]{
		data:     make([]T, capacity),
		isFull:   false,
		start:    0,
		end:      0,
		whenFull: whenFull,
		onClose:  onCloseFunc,
	}
	return rq, nil
}

func (r *unsafeRQ[T]) String() string {
	if r.closed {
		return "[RQ closed]"
	}
	return fmt.Sprintf(
		"[RQ full:%v size:%d start:%d end:%d data:%v]",
		r.isFull,
		len(r.data),
		r.start,
		r.end,
		r.data)
}

func (r *unsafeRQ[T]) Push(elem T) (int, error) {
	if r.closed {
		return 0, ErrClosed
	}
	if r.isFull {
		switch r.whenFull {
		case WhenFullError:
			return 0, ErrFullQueue
		case WhenFullOverwrite:
			// continue pushing
			break
		default:
			return 0, errors.ErrUnsupported
		}
	}

	r.data[r.end] = elem              // place the new element on the available space
	r.end = (r.end + 1) % len(r.data) // move the end forward by modulo of capacity
	r.isFull = r.end == r.start       // check if we're full now

	return r.Len(), nil
}

func (r *unsafeRQ[T]) Pop() (T, int, error) {
	var res T // "zero" element (respective of the type)
	if r.closed {
		return res, 0, ErrClosed
	}
	if !r.isFull && r.start == r.end {
		return res, 0, ErrEmptyQueue
	}

	res = r.data[r.start]                 // copy over the first element in the queue
	r.start = (r.start + 1) % len(r.data) // move the start of the queue
	r.isFull = false                      // since we're removing elements, we can never be full

	return res, r.Len(), nil
}

func (r *unsafeRQ[T]) Peek() (T, int, error) {
	var res T // "zero" element (respective of the type)
	if r.closed {
		return res, 0, ErrClosed
	}
	if !r.isFull && r.start == r.end {
		return res, 0, ErrEmptyQueue
	}
	return r.data[r.start], r.Len(), nil
}

func (r *unsafeRQ[T]) Len() int {
	if r.closed {
		return 0
	}
	res := r.end - r.start
	if res < 0 || (res == 0 && r.isFull) {
		res = len(r.data) - res
	}
	return res
}
func (r *unsafeRQ[T]) Cap() int {
	if r.closed {
		return 0
	}
	return len(r.data)
}

func (r *unsafeRQ[T]) IsFull() bool {
	if r.closed {
		return false
	}
	return r.isFull
}

func (r *unsafeRQ[T]) SetPopDeadline(t time.Time) error {
	return ErrUnsupported
}

func (r *unsafeRQ[T]) Close() error {
	r.closeOnce.Do(func() {
		r.closed = true
		if r.onClose != nil {
			r.onClose(r.data, r.start, r.end, r.isFull)
		}
		r.data = nil
	})
	return nil
}
