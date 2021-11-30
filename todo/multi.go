package todo

import (
	"math/rand"
	"sync"
)

type MultiTodo struct {
	lock  *sync.RWMutex
	todos map[int64]*Todo
}

type todo struct {
	*Todo
	k     int64
	multi *MultiTodo
}

func NewMultiTodo() *MultiTodo {
	return &MultiTodo{
		lock:  &sync.RWMutex{},
		todos: make(map[int64]*Todo),
	}
}

func (m *MultiTodo) New() *todo {
	k := rand.Int63()
	t := New()
	m.lock.Lock()
	m.todos[k] = t
	m.lock.Unlock()
	return &todo{
		t,
		k,
		m,
	}
}

func (t *todo) Close() {
	t.multi.lock.Lock()
	defer t.multi.lock.Unlock()
	delete(t.multi.todos, t.k)
}

func (m *MultiTodo) Ping() {
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, t := range m.todos {
		t.Ping()
	}
}
