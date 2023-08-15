package containerx

type Ring[T any] struct {
	size              int
	elems             []T
	head, tail, count int
}

func NewRing[T any](size int) *Ring[T] {
	if size <= 0 {
		panic("invaild ring size")
	}
	return &Ring[T]{
		elems: make([]T, size),
		size:  size,
	}
}

func (r *Ring[T]) Len() int {
	return r.count
}

func (r *Ring[T]) PushFront(elems ...T) {
	for i := range elems {
		if r.count == 0 {
			r.elems[r.head] = elems[i]
			r.head = r.prevHead()
		} else {
			r.head = r.prevHead()
			r.elems[r.head] = elems[i]
		}

		if r.count == r.size {
			r.tail = r.prevHead()
		} else {
			r.count++
		}
	}
}

func (r *Ring[T]) PushBack(elems ...T) {
	for i := range elems {
		r.elems[r.tail] = elems[i]
		r.tail = r.nextTail()

		if r.count == r.size {
			r.head = r.nextHead()
		} else {
			r.count++
		}
	}
}

func (r *Ring[T]) Front() T {
	return r.elems[r.head]
}

func (r *Ring[T]) Back() T {
	return r.elems[r.prevTail()]
}

func (r *Ring[T]) PopFront() (T, bool) {
	var t T
	if r.count <= 0 {
		return t, false
	}

	ret := r.elems[r.head]
	r.elems[r.head] = t

	r.head = r.nextHead()
	r.count--

	return ret, true
}

func (r *Ring[T]) PopBack() (T, bool) {
	var t T
	if r.count <= 0 {
		return t, false
	}

	back := r.prevTail()
	ret := r.elems[back]
	r.elems[back] = t

	r.tail = back
	r.count--

	return ret, true
}

func (r *Ring[T]) Range(f func(T)) {
	for count := 0; count < r.count; count++ {
		f(r.elems[(r.head+count)%(r.size-1)])
	}
}

func (r *Ring[T]) nextTail() int {
	return (r.tail + 1) % r.size
}

func (r *Ring[T]) prevTail() int {
	return (r.tail - 1 + r.size) % r.size
}

func (r *Ring[T]) nextHead() int {
	return (r.head + 1) % r.size
}

func (r *Ring[T]) prevHead() int {
	return (r.head - 1 + r.size) % r.size
}
