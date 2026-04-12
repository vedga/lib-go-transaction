package stack

import (
	"reflect"
	"testing"
)

func TestStack_PushAndPop(t *testing.T) {
	t.Parallel()

	s := New[int](0)

	s.Push(1)
	s.Push(2)

	val, ok := s.Pop()
	if !ok || val != 2 {
		t.Errorf("Pop() = (%v, %v), want (2, true)", val, ok)
	}

	val, ok = s.Pop()
	if !ok || val != 1 {
		t.Errorf("Pop() = (%v, %v), want (1, true)", val, ok)
	}

	val, ok = s.Pop()
	if ok {
		t.Errorf("Pop() = (%v, %v), want (_, false)", val, ok)
	}
}

func TestStack_Peek(t *testing.T) {
	t.Parallel()

	s := New[string](0)
	s.Push("first")
	s.Push("second")

	val, ok := s.Peek()
	if !ok || val != "second" {
		t.Errorf("Peek() = (%v, %v), want (\"second\", true)", val, ok)
	}

	// Peek не должен изменять стек
	size := s.Size()
	if size != 2 {
		t.Errorf("Size() = %d, want 2 (Peek должен не влиять на размер)", size)
	}
}

func TestStack_IsEmpty(t *testing.T) {
	t.Parallel()

	s := New[bool](0)

	if !s.IsEmpty() {
		t.Error("IsEmpty() = false, want true (пустой стек)")
	}

	s.Push(true)
	if s.IsEmpty() {
		t.Error("IsEmpty() = true, want false (после Push)")
	}

	s.Pop()
	if !s.IsEmpty() {
		t.Error("IsEmpty() = false, want true (после Pop до пустого)")
	}
}

func TestStack_Size(t *testing.T) {
	t.Parallel()

	s := New[int](0)

	if s.Size() != 0 {
		t.Errorf("Size() = %d, want 0", s.Size())
	}

	s.Push(1)
	s.Push(2)
	if s.Size() != 2 {
		t.Errorf("Size() = %d, want 2", s.Size())
	}

	s.Pop()
	if s.Size() != 1 {
		t.Errorf("Size() = %d, want 1", s.Size())
	}
}

func TestStack_Clear(t *testing.T) {
	t.Parallel()

	s := New[int](0)
	s.Push(1)
	s.Push(2)

	s.Clear()
	if !s.IsEmpty() {
		t.Error("Clear() не очистил стек")
	}
	if s.Size() != 0 {
		t.Errorf("Size() после Clear() = %d, want 0", s.Size())
	}

	// После Clear можно снова Push
	s.Push(42)
	val, ok := s.Pop()
	if !ok || val != 42 {
		t.Errorf("После Clear().Push(42).Pop() = (%v, %v), want (42, true)", val, ok)
	}
}

func TestStack_Values(t *testing.T) {
	t.Parallel()

	s := New[int](0)
	s.Push(1)
	s.Push(2)
	s.Push(3)

	vals := s.Values()
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(vals, want) {
		t.Errorf("Values() = %v, want %v", vals, want)
	}

	// Изменим исходный срез — это не должно повлиять на стек
	vals[0] = 999
	vals2 := s.Values()
	if !reflect.DeepEqual(vals2, want) {
		t.Errorf("Values() после изменения копии = %v, want %v", vals2, want)
	}
}

func TestStack_GenericTypes(t *testing.T) {
	t.Parallel()

	// Тест с строкой
	strStack := New[string](0)
	strStack.Push("hello")
	val, ok := strStack.Pop()
	if !ok || val != "hello" {
		t.Errorf("string: Pop() = (%v, %v), want (\"hello\", true)", val, ok)
	}

	// Тест с структурой
	type Point struct{ X, Y int }
	pointStack := New[Point](0)
	pointStack.Push(Point{10, 20})
	p, ok := pointStack.Pop()
	if !ok || p != (Point{10, 20}) {
		t.Errorf("struct: Pop() = (%v, %v), want ({10 20}, true)", p, ok)
	}

	// Тест с указателем
	ptrStack := New[*int](0)
	i := 42
	ptrStack.Push(&i)
	ptr, ok := ptrStack.Pop()
	if !ok || ptr == nil || *ptr != 42 {
		t.Errorf("*int: Pop() = (%v, %v), want (не-nil указатель на 42, true)", ptr, ok)
	}
}

func TestStack_ZeroValues(t *testing.T) {
	t.Parallel()

	s := New[int](0)

	// Пустой Pop возвращает нулевое значение типа
	val, ok := s.Pop()
	var zero int
	if val != zero || ok {
		t.Errorf("Pop() из пустого стека = (%v, %v), want (%v, false)", val, ok, zero)
	}

	// Аналогично для Peek
	val, ok = s.Peek()
	if val != zero || ok {
		t.Errorf("Peek() из пустого стека = (%v, %v), want (%v, false)", val, ok, zero)
	}
}

func TestStack_CapacityAndGrowth(t *testing.T) {
	t.Parallel()

	// Создаём с начальной ёмкостью 2
	s := New[int](2)
	if cap(s.elements) < 2 {
		t.Errorf("cap(elements) = %d, want >= 2", cap(s.elements))
	}

	// Добавляем 3 элемента — ёмкость должна вырасти
	s.Push(1)
	s.Push(2)
	s.Push(3)
	if cap(s.elements) < 3 {
		t.Errorf("cap(elements) после 3 Push = %d, want >= 3", cap(s.elements))
	}
	if s.Size() != 3 {
		t.Errorf("Size() = %d, want 3", s.Size())
	}
}
