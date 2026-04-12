package collections

type HashSet[T comparable] struct {
	data map[T]struct{}
}

func NewHashSet[T comparable]() *HashSet[T] {
	return &HashSet[T]{
		data: map[T]struct{}{},
	}
}

func (h *HashSet[T]) Add(value T) {
	h.data[value] = struct{}{}
}

func (h *HashSet[T]) Contains(value T) bool {
	_, ok := h.data[value]
	return ok
}

func (h *HashSet[T]) Remove(value T) {
	delete(h.data, value)
}

func (h *HashSet[T]) Clear() {
	h.data = map[T]struct{}{}
}

func (h *HashSet[T]) Size() int {
	return len(h.data)
}

func (h *HashSet[T]) IsEmpty() bool {
	return h.Size() == 0
}

func (h *HashSet[T]) ToSlice() []T {
	slice := make([]T, 0, len(h.data))
	for k := range h.data {
		slice = append(slice, k)
	}
	return slice
}

func (h *HashSet[T]) ToMap() map[T]struct{} {
	return h.data
}
