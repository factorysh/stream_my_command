package todo

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClose(t *testing.T) {
	m := NewMultiTodo()
	assert.Len(t, m.todos, 0)
	t1 := m.New()
	assert.Len(t, m.todos, 1)
	t1.Close()
	assert.Len(t, m.todos, 0)
}

func TestMultiTodo(t *testing.T) {
	m := NewMultiTodo()
	w := &sync.WaitGroup{}
	w.Add(3)
	t1 := m.New()
	go func() {
		for {
			<-t1.Wait()
			t1.Done()
			w.Done()
		}
	}()
	defer t1.Close()
	t2 := m.New()
	defer t2.Close()
	go func() {
		<-t2.Wait()
		w.Done()
	}()
	for i := 0; i < 3; i++ {
		m.Ping()
		fmt.Println("ping")
	}
	w.Wait()
}
