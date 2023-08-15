package containerx

import "testing"

func TestRing(t *testing.T) {
	r := NewRing[int](5)

	r.PushBack(1, 2, 3)
	if r.Len() != 3 {
		t.Errorf("expect ring len %d, but got %d", 3, r.Len())
	}

	r.PushBack(4, 5, 6)
	if r.Len() != 5 {
		t.Errorf("expect ring len %d, but got %d", 5, r.Len())
	}

	if r.Front() != 2 {
		t.Errorf("expect ring front is %d, but got %d", 2, r.Front())
	}

	if r.Back() != 6 {
		t.Errorf("expect ring back is %d, but got %d", 6, r.Back())
	}

	r.PopBack()
	got, ok := r.PopBack()
	if ok && got != 5 {
		t.Errorf("expect ring pop back is %d, but got %d", 5, got)
	}
}
