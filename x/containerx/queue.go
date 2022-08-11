package containerx

func New[T any](items ...T) *Queue[T] {
	return &Queue[T]{items}
}

type Queue[T any] struct {
	data []T
}

func (q *Queue[T]) Push(item T) {
	q.data = append(q.data, item)
}

func (q *Queue[T]) Pop() (T, bool) {
	var t T
	if q.Empty() {
		return t, false
	}

	item := q.data[0]
	q.data = q.data[1:]

	return item, true
}

func (q *Queue[T]) Front() (T, bool) {
	var t T
	if q.Empty() {
		return t, false
	}

	item := q.data[0]

	return item, true
}

func (q *Queue[T]) Back() (T, bool) {
	var t T
	if q.Empty() {
		return t, false
	}

	item := q.data[q.Len()-1]

	return item, true
}

func (q *Queue[T]) Len() int {
	return len(q.data)
}

func (q *Queue[T]) Empty() bool {
	return q.Len() == 0
}
