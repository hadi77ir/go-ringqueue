package ringqueue

import (
	"context"
	"sync"
	"time"
)

func NewSafe[T any](capacity int, whenFull WhenFull, whenEmpty WhenEmpty, onCloseFunc OnCloseFunc[T]) (RingQueue[T], error) {
	rq, err := newUnsafe[T](capacity, whenFull, onCloseFunc)
	if err != nil {
		return nil, err
	}
	if whenEmpty != WhenEmptyError && whenEmpty != WhenEmptyBlock {
		return nil, ErrUnsupported
	}
	return &safeRQ[T]{
		rq:        rq,
		available: make(chan struct{}, 1),
		deadline:  noDeadline,
		closed:    make(chan struct{}),
		whenEmpty: whenEmpty,
	}, nil
}

type safeRQ[T any] struct {
	rq    *unsafeRQ[T]
	mutex sync.Mutex

	closed    chan struct{}
	closeOnce sync.Once

	deadline      <-chan time.Time
	deadlineMutex sync.Mutex

	whenEmpty WhenEmpty
	available chan struct{}
}

// a never returning chan
var noDeadline = make(chan time.Time, 0)

func (s *safeRQ[T]) SetPopDeadline(t time.Time) error {
	if s.whenEmpty != WhenEmptyBlock {
		return ErrUnsupported
	}
	s.deadlineMutex.Lock()
	defer s.deadlineMutex.Unlock()
	if t.IsZero() {
		s.deadline = noDeadline
		return nil
	}
	s.deadline = time.After(time.Until(t))
	return nil
}

func (s *safeRQ[T]) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.closeOnce.Do(func() {
		close(s.closed)
	})
	return s.rq.Close()
}

func (s *safeRQ[T]) String() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.rq.String()
}

func (s *safeRQ[T]) Len() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.rq.Len()
}

func (s *safeRQ[T]) Cap() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.rq.Cap()
}

func (s *safeRQ[T]) guardedPush(element T) (newLen int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	newLen, err = s.rq.Push(element)
	if err != nil {
		return 0, err
	}
	return
}

func (s *safeRQ[T]) Push(element T) (newLen int, err error) {
	newLen, err = s.guardedPush(element)
	if s.whenEmpty == WhenEmptyBlock {
		select {
		case <-s.closed:
			return 0, ErrClosed
		case s.available <- struct{}{}:
			return
		default:
		}
	}
	return
}
func (s *safeRQ[T]) guardedPop() (elem T, newLen int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	elem, newLen, err = s.rq.Pop()
	return
}
func (s *safeRQ[T]) Pop() (elem T, newLen int, err error) {
	var empty T
	switch s.whenEmpty {
	case WhenEmptyError:
		return s.guardedPop()
	case WhenEmptyBlock:
		s.deadlineMutex.Lock()
		defer s.deadlineMutex.Unlock()
		select {
		case <-s.closed:
			return empty, 0, ErrClosed
		case <-s.available:
			return s.guardedPop()
		case <-s.deadline:
			return empty, 0, context.DeadlineExceeded
		}
	default:
		panic("unreachable")
	}
}

func (s *safeRQ[T]) Peek() (elem T, len int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.rq.Peek()
}
