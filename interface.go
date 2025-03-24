package ringqueue

import (
	"errors"
	"fmt"
	"io"
	"time"
)

var ErrFullQueue = fmt.Errorf("queue is full")
var ErrEmptyQueue = fmt.Errorf("queue is empty")
var ErrClosed = fmt.Errorf("queue is closed")
var ErrUnsupported = errors.ErrUnsupported

type RingQueue[T any] interface {
	fmt.Stringer
	io.Closer
	SetPopDeadline(t time.Time) error
	Len() int
	Cap() int
	Push(element T) (newLen int, err error)
	Pop() (elem T, newLen int, err error)
	Peek() (elem T, len int, err error)
}

type OnCloseFunc[T any] func(data []T, start int, end int, isFull bool)

type WhenFull int

const (
	WhenFullError = WhenFull(iota)
	WhenFullOverwrite
)

type WhenEmpty int

const (
	WhenEmptyError = WhenEmpty(iota)
	WhenEmptyBlock
)
