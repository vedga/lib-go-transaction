package queue

import (
	"testing"
)

func TestQueue_EnqueueAndDequeue(t *testing.T) {
	t.Parallel()

	q := New[int](0)

	// Добавляем элементы
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	// Проверяем размер
	if q.Size() != 3 {
		t.Errorf("Ожидаемый размер 3, получен %d", q.Size())
	}

	// Извлекаем элементы в порядке FIFO
	val1, ok1 := q.Dequeue()
	if !ok1 || val1 != 1 {
		t.Errorf("Первый Dequeue: ожидаем (1, true), получили (%v, %v)", val1, ok1)
	}

	val2, ok2 := q.Dequeue()
	if !ok2 || val2 != 2 {
		t.Errorf("Второй Dequeue: ожидаем (2, true), получили (%v, %v)", val2, ok2)
	}

	val3, ok3 := q.Dequeue()
	if !ok3 || val3 != 3 {
		t.Errorf("Третий Dequeue: ожидаем (3, true), получили (%v, %v)", val3, ok3)
	}

	// Очередь должна быть пустой
	if !q.IsEmpty() {
		t.Error("Очередь должна быть пустой после извлечения всех элементов")
	}
}

func TestQueue_DequeueEmpty(t *testing.T) {
	t.Parallel()

	q := New[string](0)

	_, ok := q.Dequeue()
	if ok {
		t.Error("Dequeue из пустой очереди должен возвращать false")
	}
}

func TestQueue_Peek(t *testing.T) {
	t.Parallel()

	q := New[float64](0)
	q.Enqueue(1.1)
	q.Enqueue(2.2)

	val, ok := q.Peek()
	if !ok || val != 1.1 {
		t.Errorf("Peek должен вернуть первый элемент: ожидаем (1.1, true), получили (%v, %v)", val, ok)
	}

	// Peek не должен удалять элемент
	if q.Size() != 2 {
		t.Error("Peek не должен изменять размер очереди")
	}
}

func TestQueue_IsEmpty(t *testing.T) {
	t.Parallel()

	q := New[bool](0)

	if !q.IsEmpty() {
		t.Error("Новая очередь должна быть пустой")
	}

	q.Enqueue(true)
	if q.IsEmpty() {
		t.Error("После Enqueue очередь не должна быть пустой")
	}

	q.Dequeue()
	if !q.IsEmpty() {
		t.Error("После извлечения всех элементов очередь должна быть пустой")
	}
}

func TestQueue_Size(t *testing.T) {
	t.Parallel()

	q := New[int](0)

	if q.Size() != 0 {
		t.Errorf("Размер новой очереди должен быть 0, получен %d", q.Size())
	}

	q.Enqueue(1)
	q.Enqueue(2)
	if q.Size() != 2 {
		t.Errorf("Размер после двух Enqueue должен быть 2, получен %d", q.Size())
	}

	q.Dequeue()
	if q.Size() != 1 {
		t.Errorf("Размер после одного Dequeue должен быть 1, получен %d", q.Size())
	}
}

func TestQueue_Clear(t *testing.T) {
	t.Parallel()

	q := New[string](0)
	q.Enqueue("a")
	q.Enqueue("b")

	q.Clear()
	if !q.IsEmpty() || q.Size() != 0 {
		t.Error("После Clear очередь должна быть пустой и иметь размер 0")
	}
}

func TestQueue_ToSlice(t *testing.T) {
	t.Parallel()

	q := New[int](0)
	q.Enqueue(10)
	q.Enqueue(20)
	q.Enqueue(30)

	slice := q.ToSlice()
	expected := []int{10, 20, 30}

	if len(slice) != len(expected) {
		t.Errorf("Длина слайса должна быть %d, получена %d", len(expected), len(slice))
	}

	for i, v := range expected {
		if slice[i] != v {
			t.Errorf("Элемент %d: ожидаем %d, получен %d", i, v, slice[i])
		}
	}

	// Проверяем, что ToSlice не изменяет очередь
	if q.Size() != 3 {
		t.Error("ToSlice не должен изменять состояние очереди")
	}
}

func TestQueue_GenericTypes(t *testing.T) {
	t.Parallel()

	// Тестируем с разными типами

	// Строки
	stringQ := New[string](0)
	stringQ.Enqueue("hello")
	stringQ.Enqueue("world")
	s1, _ := stringQ.Dequeue()
	s2, _ := stringQ.Dequeue()
	if s1 != "hello" || s2 != "world" {
		t.Error("Очередь строк работает некорректно")
	}

	// Структуры
	type Point struct{ X, Y int }
	pointQ := New[Point](0)
	pointQ.Enqueue(Point{1, 2})
	pointQ.Enqueue(Point{3, 4})
	p1, _ := pointQ.Dequeue()
	p2, _ := pointQ.Dequeue()
	if p1.X != 1 || p1.Y != 2 || p2.X != 3 || p2.Y != 4 {
		t.Error("Очередь структур работает некорректно")
	}
}
