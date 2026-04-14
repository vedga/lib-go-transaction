package deque

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
)

// Deque — generic-очередь с доступом с обеих сторон (double-ended queue)
type Deque[T any] struct {
	items []T
}

// New создаёт новую пустую двустороннюю очередь
func New[T any](capacity int) *Deque[T] {
	return &Deque[T]{
		items: make([]T, 0, capacity),
	}
}

// MarshalJSON implementation of json.Marshaler interface
func (d *Deque[T]) MarshalJSON() ([]byte, error) {
	jsonData, e := json.Marshal(d.items)
	if e != nil {
		return nil, fmt.Errorf(`encode deque error: %w`, e)
	}

	var compact bytes.Buffer
	if e = json.Compact(&compact, jsonData); e != nil {
		return nil, fmt.Errorf(`compact deque error: %w`, e)
	}

	return compact.Bytes(), nil
}

// UnmarshalJSON implementation json.Unmarshaler interface
func (d *Deque[T]) UnmarshalJSON(raw []byte) error {
	items := make([]T, 0)
	if e := json.Unmarshal(raw, &items); e != nil {
		return fmt.Errorf("stack unmarshalling error: %w", e)
	}

	d.items = items

	return nil
}

// PushFront добавляет элемент в начало очереди
func (d *Deque[T]) PushFront(item T) {
	d.items = append([]T{item}, d.items...)
}

// PushBack добавляет элемент в конец очереди
func (d *Deque[T]) PushBack(item T) {
	d.items = append(d.items, item)
}

// PopFront удаляет и возвращает элемент из начала очереди
// Возвращает значение и флаг успеха (false, если очередь пуста)
func (d *Deque[T]) PopFront() (T, bool) {
	if d.IsEmpty() {
		var zero T
		return zero, false
	}

	item := d.items[0]
	d.items = slices.Delete(d.items, 0, 1)
	return item, true
}

// PopBack удаляет и возвращает элемент из конца очереди
// Возвращает значение и флаг успеха
func (d *Deque[T]) PopBack() (T, bool) {
	if d.IsEmpty() {
		var zero T
		return zero, false
	}

	lastIndex := len(d.items) - 1
	item := d.items[lastIndex]
	d.items = slices.Delete(d.items, lastIndex, lastIndex+1)
	return item, true
}

// PeekFront возвращает элемент из начала очереди без удаления
// Возвращает значение и флаг успеха
func (d *Deque[T]) PeekFront() (T, bool) {
	if d.IsEmpty() {
		var zero T
		return zero, false
	}
	return d.items[0], true
}

// PeekBack возвращает элемент из конца очереди без удаления
// Возвращает значение и флаг успеха
func (d *Deque[T]) PeekBack() (T, bool) {
	if d.IsEmpty() {
		var zero T
		return zero, false
	}
	lastIndex := len(d.items) - 1
	return d.items[lastIndex], true
}

// IsEmpty проверяет, пуста ли очередь
func (d *Deque[T]) IsEmpty() bool {
	return len(d.items) == 0
}

// Size возвращает количество элементов в очереди
func (d *Deque[T]) Size() int {
	return len(d.items)
}

// Clear очищает очередь
func (d *Deque[T]) Clear() {
	d.items = make([]T, 0, cap(d.items))
}

// ToSlice возвращает копию элементов очереди в виде слайса
func (d *Deque[T]) ToSlice() []T {
	result := make([]T, len(d.items))
	copy(result, d.items)
	return result
}
