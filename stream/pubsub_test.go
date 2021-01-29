package stream

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockWriter struct {
	b  *bytes.Buffer
	wg *sync.WaitGroup
	i  int
}

func (m *mockWriter) Write(buff []byte) (int, error) {
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	fmt.Println(m.i, "write")
	defer m.wg.Done()
	return m.b.Write(buff)
}

func TestOnePublisher(t *testing.T) {
	p, err := NewPubsub(os.TempDir())
	assert.NoError(t, err)
	defer os.Remove(p.buffer.Name())

	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)

		ctx := context.TODO()
		err = p.Subscribe(ctx, 0, &mockWriter{
			b:  &bytes.Buffer{},
			wg: wg,
			i:  i,
		})
		assert.NoError(t, err)
		fmt.Println("subscriber", i)
	}
	_, err = p.Write([]byte("beuha"))
	assert.NoError(t, err)
	fmt.Println("Wrote")
	wg.Wait()
}
