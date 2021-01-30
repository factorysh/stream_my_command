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

	READERS := 10
	BLOBS := 10

	wg := &sync.WaitGroup{}
	wg.Add(READERS * BLOBS)

	rs := make([]*mockWriter, READERS)
	for i := 0; i < READERS; i++ {
		ctx := context.TODO()
		r := &mockWriter{
			b:  &bytes.Buffer{},
			wg: wg,
			i:  i,
		}
		err = p.Subscribe(ctx, 0, r)
		assert.NoError(t, err)
		rs[i] = r
		fmt.Println("subscriber", i)
	}
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for i := 0; i < BLOBS; i++ {
		buff := make([]byte, 1024*1024)
		_, err := r.Read(buff)
		assert.NoError(t, err)
		_, err = p.Write(buff)
		assert.NoError(t, err)
		fmt.Println("Wrote")
	}
	wg.Wait()
	for _, r := range rs {
		assert.Equal(t, BLOBS*1024*1024, r.b.Len())
	}

}
