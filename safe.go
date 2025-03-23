package ringqueue

import "sync"

func NewSafe[T any](capacity int, whenFull WhenFull) (RingQueue[T], error) {
	rq, err := newUnsafe[T](capacity, WhenFullError)
	if err != nil {
		return nil, err
	}
	return &safeRQ[T]{
		rq: rq,
	}, nil
}

type safeRQ[T any] struct {
	rq    *unsafeRQ[T]
	mutex sync.Mutex
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

func (s *safeRQ[T]) Push(element T) (newLen int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.rq.Push(element)
}

func (s *safeRQ[T]) Pop() (elem T, newLen int, ok bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.rq.Pop()
}

func (s *safeRQ[T]) Peek() (elem T, len int, ok bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.rq.Peek()
}
