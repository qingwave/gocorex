package containerx

import "testing"

func TestHeap(t *testing.T) {
	data := []int{3, 1, 2}
	h := NewHeap(data, func(x, y int) bool {
		return x < y
	})

	old := h.Len()
	if h.Len() != len(data) {
		t.Errorf("expected len %d, got %d", len(data), h.Len())
	}

	top, ok := h.Peek()
	if !ok || top != 1 {
		t.Errorf("peek item unexpected, item: %v, ok: %t", top, ok)
	}

	item, ok := h.Pop()
	if !ok || top != 1 {
		t.Errorf("pop item unexpected, item: %v, ok: %t", item, ok)
	}

	h.Push(4)
	if h.Len() != old {
		t.Errorf("expected len %d, got %d", old, h.Len())
	}
}
