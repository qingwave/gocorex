package containerx

type Heap[T any] struct {
	data []T
	less func(x, y T) bool
}

func NewHeap[T any](data []T, less func(x, y T) bool) *Heap[T] {
	h := &Heap[T]{data: data, less: less}
	n := h.Len()
	for i := n/2 - 1; i >= 0; i-- {
		h.down(i, n)
	}
	return h
}

func (h *Heap[T]) Push(x T) {
	h.data = append(h.data, x)
	h.up(h.Len() - 1)
}

func (h *Heap[T]) Pop() (T, bool) {
	var t T
	if h.Empty() {
		return t, false
	}

	n := h.Len() - 1
	h.Swap(0, n)
	h.down(0, n)

	item := h.data[h.Len()-1]
	h.data[h.Len()-1] = t // set zero value
	h.data = h.data[0 : h.Len()-1]
	return item, true
}

func (h *Heap[T]) Peek() (T, bool) {
	if h.Empty() {
		var t T
		return t, false
	}

	return h.data[0], true
}

func (h *Heap[T]) Remove(i int) T {
	if i >= h.Len() {
		var t T
		return t
	}

	n := h.Len() - 1
	if n != i {
		h.Swap(i, n)
		if !h.down(i, n) {
			h.up(i)
		}
	}

	t, _ := h.Pop()
	return t
}

func (h *Heap[T]) Fix(i int) {
	if h.Empty() {
		return
	}

	if !h.down(i, h.Len()) {
		h.up(i)
	}
}

func (h *Heap[T]) Len() int {
	return len(h.data)
}

func (h *Heap[T]) Empty() bool {
	return h.Len() == 0
}

func (h *Heap[T]) Less(i, j int) bool {
	return h.less(h.data[i], h.data[j])
}

func (h *Heap[T]) Swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}

func (h *Heap[T]) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		j = i
	}
}

func (h *Heap[T]) down(i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && h.Less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		if !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		i = j
	}
	return i > i0
}
