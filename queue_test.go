package ringqueue

import (
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
	obj, _ := newUnsafe[int](10, WhenFullError)
	expected := "[RQ full:false size:10 start:0 end:0 data:[0 0 0 0 0 0 0 0 0 0]]"
	actual := fmt.Sprint(obj)

	if actual != expected {
		t.Fatalf("Mismatch, expected:%s, found:%s", expected, actual)
	}
}

func TestPushEnough(t *testing.T) {
	obj, _ := newUnsafe[int](10, WhenFullError)
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
	obj, _ := newUnsafe[int](10, WhenFullError)
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
	obj, _ := newUnsafe[int](10, WhenFullError)
	for idx := 0; idx < 8; idx++ {
		obj.Push(idx)
	}
	for idx := 0; idx < 5; idx++ {
		e, _, ok := obj.Pop()
		if ok || e != idx {
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
	rr, _ := newUnsafe[int](capacity, WhenFullError)

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

	fmt.Println("rr")
	cap = 1
	for idx := 1; idx < 7; idx++ {
		simRQ(cap)
		cap = cap * 10
	}
}

func BenchmarkRQUnsafe(b *testing.B) {
	rq, _ := newUnsafe[int](1_000, WhenFullOverwrite)

	for n := 0; n < b.N; n++ {
		rq.Push(n)
	}
}

func BenchmarkRQSafe(b *testing.B) {
	rq, _ := NewSafe[int](1_000, WhenFullOverwrite)

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
