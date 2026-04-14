package deque

import (
	"reflect"
	"testing"
)

func TestDeque_Clear(t *testing.T) {
	t.Parallel()

	d := New[int](0)
	d.PushBack(1)
	d.PushBack(2)
	d.PushBack(3)

	if d.Size() != 3 {
		t.Error("Перед очисткой размер должен быть 3")
	}

	d.Clear()

	if !d.IsEmpty() || d.Size() != 0 {
		t.Error("После Clear очередь должна быть пустой и иметь размер 0")
	}

	// Проверяем, что после очистки можно добавлять новые элементы
	d.PushFront(42)
	val, ok := d.PopFront()
	if !ok || val != 42 {
		t.Errorf("После очистки и добавления элемента: ожидаем (42, true), получили (%v, %v)", val, ok)
	}
}

func TestDeque_EmptyOperations(t *testing.T) {
	t.Parallel()

	d := New[string](0)

	// Все операции с пустой очередью
	_, ok1 := d.PopFront()
	_, ok2 := d.PopBack()
	_, ok3 := d.PeekFront()
	_, ok4 := d.PeekBack()

	if ok1 || ok2 || ok3 || ok4 {
		t.Error("Все операции с пустой очередью должны возвращать false во втором значении")
	}

	// Zero value для разных типов
	intDeque := New[int](0)
	valInt, _ := intDeque.PopFront()
	if valInt != 0 {
		t.Errorf("Zero value для int должно быть 0, получено: %v", valInt)
	}

	boolDeque := New[bool](0)
	valBool, _ := boolDeque.PopFront()
	if valBool != false {
		t.Errorf("Zero value для bool должно быть false, получено: %v", valBool)
	}

	stringDeque := New[string](0)
	valString, _ := stringDeque.PopFront()
	if valString != "" {
		t.Errorf("Zero value для string должно быть \"\", получено: %q", valString)
	}
}

func TestDeque_LargeNumberOfElements(t *testing.T) {
	t.Parallel()

	d := New[int](0)
	n := 1000

	// Добавляем много элементов
	for i := 0; i < n; i++ {
		d.PushBack(i)
	}

	if d.Size() != n {
		t.Errorf("После добавления %d элементов размер должен быть %d, получен: %d", n, n, d.Size())
	}

	// Извлекаем все элементы
	for i := 0; i < n; i++ {
		val, ok := d.PopFront()
		if !ok || val != i {
			t.Errorf("Элемент %d: ожидаем (%d, true), получили (%v, %v)", i, i, val, ok)
		}
	}

	if !d.IsEmpty() {
		t.Error("После извлечения всех элементов очередь должна быть пустой")
	}
}

func TestDeque_MixedTypes(t *testing.T) {
	t.Parallel()
	// Тестируем с разными типами

	// Строки
	stringDeque := New[string](0)
	stringDeque.PushFront("first")
	stringDeque.PushBack("last")

	frontStr, _ := stringDeque.PeekFront()
	backStr, _ := stringDeque.PeekBack()

	if frontStr != "first" || backStr != "last" {
		t.Errorf("Строки: ожидаем front=\"first\", back=\"last\", получили front=%q, back=%q", frontStr, backStr)
	}

	// Структуры
	type Point struct {
		X, Y int
	}

	pointDeque := New[Point](0)
	p1 := Point{X: 1, Y: 2}
	p2 := Point{X: 3, Y: 4}

	pointDeque.PushFront(p1)
	pointDeque.PushBack(p2)

	frontPoint, _ := pointDeque.PeekFront()
	backPoint, _ := pointDeque.PeekBack()

	if frontPoint != p1 || backPoint != p2 {
		t.Errorf("Структуры: неверные значения при извлечении")
	}
}

func TestDeque_ToSlice(t *testing.T) {
	t.Parallel()

	d := New[int](0)

	// Пустая очередь
	emptySlice := d.ToSlice()
	if len(emptySlice) != 0 {
		t.Error("ToSlice для пустой очереди должен возвращать пустой слайс")
	}

	// Очередь с элементами
	d.PushBack(1)
	d.PushBack(2)
	d.PushBack(3)

	slice := d.ToSlice()

	// Проверяем, что срез — независимая копия
	slice[0] = 999
	first, _ := d.PeekFront()
	if first != 1 {
		t.Error("Изменение возвращённого среза не должно влиять на очередь")
	}

	expected := []int{1, 2, 3}
	actual := d.ToSlice() // берём ещё одну копию

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("ToSlice: ожидаем %v, получили %v", expected, actual)
	}
}
