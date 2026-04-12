package queue

import (
	"encoding/json"
	"fmt"
	"slices"
)

// Queue — generic-очередь FIFO на основе слайса
type Queue[T any] struct {
	items []T
}

// New создаёт новую пустую очередь
func New[T any](capacity int) *Queue[T] {
	return &Queue[T]{
		items: make([]T, capacity),
	}
}

// MarshalJSON implementation of json.Marshaler interface
func (q *Queue[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(q.ToSlice())
}

// UnmarshalJSON implementation json.Unmarshaler interface
func (q *Queue[T]) UnmarshalJSON(raw []byte) error {
	elements := make([]T, 0)
	if e := json.Unmarshal(raw, &elements); e != nil {
		return fmt.Errorf("stack unmarshalling error: %w", e)
	}

	q.items = elements

	return nil
}

// Enqueue добавляет элемент в конец очереди
func (q *Queue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

// Dequeue удаляет и возвращает элемент из начала очереди
// Возвращает значение и флаг успеха (false, если очередь пуста)
func (q *Queue[T]) Dequeue() (T, bool) {
	if q.IsEmpty() {
		var zero T
		return zero, false
	}

	item := q.items[0]

	// Remove most element
	return item, q.Drop()
}

// Drop most element
func (q *Queue[T]) Drop() bool {
	if q.IsEmpty() {
		return false
	}

	// Используем slices.Delete для удаления первого элемента
	q.items = slices.Delete(q.items, 0, 1)

	return true
}

// Peek возвращает элемент из начала очереди без удаления
// Возвращает значение и флаг успеха
func (q *Queue[T]) Peek() (T, bool) {
	if q.IsEmpty() {
		var zero T
		return zero, false
	}
	return q.items[0], true
}

// IsEmpty проверяет, пуста ли очередь
func (q *Queue[T]) IsEmpty() bool {
	return len(q.items) == 0
}

// Size возвращает количество элементов в очереди
func (q *Queue[T]) Size() int {
	return len(q.items)
}

// Clear очищает очередь
func (q *Queue[T]) Clear() {
	q.items = make([]T, 0)
}

// ToSlice возвращает копию элементов очереди в виде слайса
func (q *Queue[T]) ToSlice() []T {
	result := make([]T, len(q.items))
	copy(result, q.items)
	return result
}
