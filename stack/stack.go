package stack

import (
	"encoding/json"
	"fmt"
)

// Stack — универсальный стек для любого типа T.
type Stack[T any] struct {
	elements []T
}

// New создаёт новый стек с заданной начальной ёмкостью.
// capacity=0 — без предварительного выделения памяти.
func New[T any](capacity int) *Stack[T] {
	return &Stack[T]{
		elements: make([]T, 0, capacity),
	}
}

// MarshalJSON implementation of json.Marshaler interface
func (s *Stack[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.elements)
}

// UnmarshalJSON implementation json.Unmarshaler interface
func (s *Stack[T]) UnmarshalJSON(raw []byte) error {
	elements := make([]T, 0)
	if e := json.Unmarshal(raw, &elements); e != nil {
		return fmt.Errorf("stack unmarshalling error: %w", e)
	}

	s.elements = elements

	return nil
}

// Push добавляет элемент в вершину стека.
func (s *Stack[T]) Push(values ...T) {
	s.elements = append(s.elements, values...)
}

// Pop удаляет и возвращает верхний элемент.
// Возвращает: (значение, true) — если элемент есть; (нулевое значение, false) — если стек пуст.
func (s *Stack[T]) Pop() (T, bool) {
	n := len(s.elements)
	if n == 0 {
		var zero T
		return zero, false
	}

	last := s.elements[n-1]
	s.elements = s.elements[:n-1] // перерезаем срез, не освобождая память
	return last, true
}

// Peek возвращает верхний элемент без удаления.
// Аналогично Pop(), но не изменяет стек.
func (s *Stack[T]) Peek() (T, bool) {
	n := len(s.elements)
	if n == 0 {
		var zero T
		return zero, false
	}

	return s.elements[n-1], true
}

// IsEmpty проверяет, пуст ли стек.
func (s *Stack[T]) IsEmpty() bool {
	return len(s.elements) == 0
}

// Size возвращает количество элементов в стеке.
func (s *Stack[T]) Size() int {
	return len(s.elements)
}

// Clear удаляет все элементы из стека.
func (s *Stack[T]) Clear() {
	s.elements = nil // освобождаем ссылку, память может быть собрана GC
}

// Values возвращает срез всех элементов стека (от дна к вершине).
// Не изменяет стек.
func (s *Stack[T]) Values() []T {
	return append([]T(nil), s.elements...) // копируем, чтобы избежать внешних модификаций
}
