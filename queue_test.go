package ringqueue

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func eqSlices[T comparable](a []T, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for idx := 0; idx < len(a); idx++ {
		if a[idx] != b[idx] {
			return false
		}
	}

	return true
}

func TestToString(t *testing.T) {
	obj, _ := newUnsafe[int](10, WhenFullError, nil)
	expected := "[RQ full:false size:10 start:0 end:0 data:[0 0 0 0 0 0 0 0 0 0]]"
	actual := fmt.Sprint(obj)

	if actual != expected {
		t.Fatalf("Mismatch, expected:%s, found:%s", expected, actual)
	}
}

func TestPushEnough(t *testing.T) {
	obj, _ := newUnsafe[int](10, WhenFullError, nil)
	for idx := 0; idx < 10; idx++ {
		_, err := obj.Push(idx)
		if err != nil {
			t.Fatalf("Unexpected error in adding an element with index %d", idx)
		}
	}

	expected := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	if !eqSlices(obj.data, expected) {
		t.Fatalf("Container data mismatch, expected:%v, found:%v", expected, obj.data)
	}
}

func TestPushOver(t *testing.T) {
	obj, _ := newUnsafe[int](10, WhenFullError, nil)
	for idx := 0; idx < 10; idx++ {
		_, err := obj.Push(idx)
		if err != nil {
			t.Fatalf("Unexpected error in adding an element with index %d", idx)
		}
	}

	_, err := obj.Push(100)
	if err == nil {
		t.Fatalf("Expected overflow error")
	}

	expected := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	if !eqSlices(obj.data, expected) {
		t.Fatalf("Container data mismatch, expected:%v, found:%v", expected, obj.data)
	}
}

func TestPushPop(t *testing.T) {
	obj, _ := newUnsafe[int](10, WhenFullError, nil)
	for idx := 0; idx < 8; idx++ {
		obj.Push(idx)
	}
	for idx := 0; idx < 5; idx++ {
		e, _, err := obj.Pop()
		if err != nil || e != idx {
			t.Fatalf("inconsistent behavior")
		}
	}
	for idx := 0; idx < 7; idx++ {
		obj.Push(100 + idx)
	}

	expected := []int{102, 103, 104, 105, 106, 5, 6, 7, 100, 101}

	if !eqSlices(obj.data, expected) {
		t.Fatalf("Container data mismatch, expected:%v, found:%v", expected, obj.data)
	}

	if obj.Len() != 10 {
		t.Fatalf("inconsistent size: %d", obj.Len())
	}

	for idx := 0; idx < 10; idx++ {
		e, _, _ := obj.Pop()
		if e != expected[(5+idx)%10] {
			t.Fatalf("inconsistent behavior")
		}
	}
}

func TestDeadline(t *testing.T) {
	obj, _ := NewSafe[int](10, WhenFullError, WhenEmptyBlock, nil)
	timeBefore := time.Now()
	obj.SetPopDeadline(time.Now().Add(1 * time.Second))
	_, _, err := obj.Pop()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected DeadlineExceeded error")
	}
	if time.Since(timeBefore) < 1*time.Second {
		t.Fatalf("Expected 1s timeout")
	}
}

func TestDeadline2(t *testing.T) {
	obj, _ := NewSafe[int](10, WhenFullError, WhenEmptyBlock, nil)
	timeBefore := time.Now()
	for i := 0; i < 10; i++ {
		t.Log("push ", i)
		obj.Push(i)
	}
	for i := 0; i < 10; i++ {
		t.Log("pop ", i)
		obj.Pop()
	}
	obj.SetPopDeadline(time.Now().Add(5 * time.Second))
	_, _, err := obj.Pop()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected DeadlineExceeded error")
	}
	if time.Since(timeBefore) < 5*time.Second {
		t.Fatalf("Expected 1s timeout")
	}
}

func sim(capacity int) {
	ar := make([]int, capacity, capacity)
	size := 0

	start := time.Now()
	for n := 0; n < 1000000; n++ {
		if size >= len(ar) {
			copy(ar[0:], ar[1:])
			size--
		}

		ar[size] = n
		size++
	}

	fmt.Printf("%d took %v\n", capacity, time.Since(start).Seconds())
}

func simRQ(capacity int) {
	rr, _ := newUnsafe[int](capacity, WhenFullError, nil)

	start := time.Now()
	for n := 0; n < 1000000; n++ {
		if rr.IsFull() {
			rr.Pop()
		}
		rr.Push(n)
	}

	fmt.Printf("%d took %v\n", capacity, time.Since(start).Seconds())
}

func TestSizes(t *testing.T) {
	fmt.Println("array")
	cap := 1
	for idx := 1; idx < 7; idx++ {
		sim(cap)
		cap = cap * 10
	}

	fmt.Println("rq")
	cap = 1
	for idx := 1; idx < 7; idx++ {
		simRQ(cap)
		cap = cap * 10
	}
}

func BenchmarkRQUnsafe(b *testing.B) {
	rq, _ := newUnsafe[int](1_000, WhenFullOverwrite, nil)

	for n := 0; n < b.N; n++ {
		rq.Push(n)
	}
}

func BenchmarkRQSafe(b *testing.B) {
	rq, _ := NewSafe[int](1_000, WhenFullOverwrite, WhenEmptyError, nil)

	for n := 0; n < b.N; n++ {
		rq.Push(n)
	}
}

func BenchmarkArray(b *testing.B) {
	var ar [1_000]int
	size := 0

	for n := 0; n < b.N; n++ {
		if size >= len(ar) {
			copy(ar[0:], ar[1:])
			size--
		}

		ar[size] = n
		size++
	}
}

func TestOnClose(t *testing.T) {
	type testCase struct {
		name           string
		pushCount      int
		popCount       int
		onCloseCount   int
		wantErrInClose bool
		wantMismatch   bool
	}
	tests := []testCase{
		{
			name:           "under-push",
			pushCount:      1,
			popCount:       1,
			onCloseCount:   0,
			wantErrInClose: false,
			wantMismatch:   false,
		},
		{
			name:           "over-push",
			pushCount:      15,
			popCount:       10,
			onCloseCount:   0,
			wantErrInClose: false,
			wantMismatch:   false,
		},
		{
			name:           "over-push and under-pop",
			pushCount:      15,
			popCount:       5,
			onCloseCount:   5,
			wantErrInClose: false,
			wantMismatch:   false,
		},
		{
			name:           "under-push and under-pop",
			pushCount:      7,
			popCount:       5,
			onCloseCount:   2,
			wantErrInClose: false,
			wantMismatch:   false,
		},
		{
			name:           "over-push and over-pop",
			pushCount:      15,
			popCount:       15,
			onCloseCount:   0,
			wantErrInClose: false,
			wantMismatch:   false,
		},
		{
			name:           "mismatch",
			pushCount:      10,
			popCount:       0,
			onCloseCount:   0,
			wantErrInClose: false,
			wantMismatch:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			onCloseCount := 0
			rr, _ := newUnsafe[int](10, WhenFullError, func(data int) {
				onCloseCount++
			})
			for i := 0; i < tt.pushCount; i++ {
				rr.Push(i)
			}
			for i := 0; i < tt.popCount; i++ {
				rr.Pop()
			}
			if err := rr.Close(); (err != nil) != tt.wantErrInClose {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErrInClose)
			}
			if onCloseCount != tt.onCloseCount && !tt.wantMismatch {
				t.Errorf("onCloseCount = %v, wanted onCloseCount %v", onCloseCount, tt.onCloseCount)
			}
		})
	}
}
