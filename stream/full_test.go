package stream

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	l, err := NewBucket(os.TempDir(), 10*1024*1024)
	assert.NoError(t, err)
	assert.NotEmpty(t, l.ID())

	r := rand.New(rand.NewSource(time.Now().Unix()))
	buff := make([]byte, 3*1024*1024)
	h1 := sha256.New()
	for i := 0; i < 10; i++ {
		_, _ = r.Read(buff)
		n, err := l.Write(buff)
		assert.NoError(t, err)
		_, err = h1.Write(buff)
		assert.NoError(t, err)
		assert.Equal(t, 3*1024*1024, n)
		fmt.Println("Write", i)
		time.Sleep(time.Duration(r.Int63n(10)) * time.Millisecond)
	}
	buff = make([]byte, 512)
	_, _ = r.Read(buff)
	w, err := l.Write(buff)
	assert.NoError(t, err)
	_, err = h1.Write(buff)
	assert.NoError(t, err)
	assert.Equal(t, 512, w)

	assert.Equal(t, 30*1024*1024+512, l.Len())
	err = l.Close()
	assert.NoError(t, err)
	assert.True(t, l.Closed())
	out := bytes.NewBuffer(nil)
	err = l.Copy(0, out)
	assert.NoError(t, err)
	fmt.Println("path", l.Path())
	assert.Equal(t, 30*1024*1024+512, out.Len())
	h2 := sha256.New()
	h2.Write(out.Bytes())
	assert.Equal(t, h1.Sum(nil), l.Hash())
}

func TestReadWhileWrite(t *testing.T) {
	l, err := NewBucket(os.TempDir(), 10*1024*1024)
	assert.NoError(t, err)

	r := rand.New(rand.NewSource(time.Now().Unix()))
	buff := make([]byte, 3*1024*1024)
	h1 := sha256.New()
	for i := 0; i < 5; i++ {
		_, _ = r.Read(buff)
		n, err := l.Write(buff)
		assert.NoError(t, err)
		_, err = h1.Write(buff)
		assert.NoError(t, err)
		assert.Equal(t, 3*1024*1024, n)
		fmt.Println("Write", i)
		time.Sleep(time.Duration(r.Int63n(10)) * time.Millisecond)
	}
	wg := &sync.WaitGroup{}

	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func() {
			b := bytes.NewBuffer(nil)
			err = l.Copy(0, b)
			assert.NoError(t, err)
			fmt.Println("path", l.Path())
			assert.Equal(t, 30*1024*1024+512, b.Len())
			h2 := sha256.New()
			h2.Write(b.Bytes())
			assert.Equal(t, h1.Sum(nil), l.Hash())
			wg.Done()
		}()
	}
	time.Sleep(10 * time.Millisecond)

	go func() {
		for i := 0; i < 5; i++ {
			_, _ = r.Read(buff)
			_, err := l.Write(buff)
			assert.NoError(t, err)
			_, err = h1.Write(buff)
			assert.NoError(t, err)
			fmt.Println("Write", i)
			time.Sleep(time.Duration(r.Int63n(10)) * time.Millisecond)
		}

		buff = make([]byte, 512)
		_, _ = r.Read(buff)
		w, err := l.Write(buff)
		assert.NoError(t, err)
		_, err = h1.Write(buff)
		assert.NoError(t, err)
		assert.Equal(t, 512, w)

		assert.Equal(t, 30*1024*1024+512, l.Len())
		err = l.Close()
		assert.NoError(t, err)
	}()

	wg.Wait()

	// Read a finished lb
	b := bytes.NewBuffer(nil)
	err = l.Copy(0, b)
	assert.NoError(t, err)
	fmt.Println("path", l.Path())
	assert.Equal(t, 30*1024*1024+512, b.Len())
	h2 := sha256.New()
	h2.Write(b.Bytes())
	assert.Equal(t, h1.Sum(nil), l.Hash())
}

type SlowWriter struct {
	b *bytes.Buffer
}

func (s *SlowWriter) Write(b []byte) (int, error) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	time.Sleep(time.Duration(r.Int63n(30)) * time.Millisecond)
	return s.b.Write(b)
}

func TestLongBuffer(t *testing.T) {
	l, err := NewBucket(os.TempDir(), 10*1024*1024)
	assert.NoError(t, err)

	r := rand.New(rand.NewSource(time.Now().Unix()))
	wg := sync.WaitGroup{}
	wg.Add(2)

	h3 := sha256.New()
	go func() { // Slow reader
		b := &SlowWriter{bytes.NewBuffer(nil)}
		err = l.Copy(0, b)
		assert.NoError(t, err)
		assert.Equal(t, 30*1024*1024, b.b.Len())
		h3.Write(b.b.Bytes())
		wg.Done()
	}()

	h4 := sha256.New()
	go func() {
		b := bytes.NewBuffer(nil)
		err = l.Copy(0, b)
		assert.NoError(t, err)
		assert.Equal(t, 30*1024*1024, b.Len())
		h4.Write(b.Bytes())
		wg.Done()
	}()

	buff := make([]byte, 3*1024*1024)
	h1 := sha256.New()
	for i := 0; i < 10; i++ {
		_, _ = r.Read(buff)
		n, err := l.Write(buff)
		assert.NoError(t, err)
		_, err = h1.Write(buff)
		assert.NoError(t, err)
		assert.Equal(t, 3*1024*1024, n)
		fmt.Println("Write", i)
		time.Sleep(time.Duration(r.Int63n(10)) * time.Millisecond)
	}
	assert.Equal(t, 30*1024*1024, l.Len())
	l.Close()

	wg.Wait()

	b := bytes.NewBuffer(nil)
	err = l.Copy(0, b)
	assert.NoError(t, err)
	fmt.Println("path", l.Path())
	assert.Equal(t, 30*1024*1024, b.Len())
	h2 := sha256.New()
	h2.Write(b.Bytes())
	assert.Equal(t, l.Hash(), h1.Sum(nil))
	assert.Equal(t, l.Hash(), h2.Sum(nil))
	assert.Equal(t, l.Hash(), h3.Sum(nil))
	assert.Equal(t, l.Hash(), h4.Sum(nil))
}

func TestSeek(t *testing.T) {
	l, err := NewBucket(os.TempDir(), 10*1024*1024)
	assert.NoError(t, err)
	r := rand.New(rand.NewSource(time.Now().Unix()))
	buff := make([]byte, 3*1024*1024)
	start := bytes.NewBuffer(nil)
	SEEK := 5 * 1024 * 1024
	for i := 0; i < 10; i++ {
		_, _ = r.Read(buff)
		n, err := l.Write(buff)
		assert.NoError(t, err)
		assert.Equal(t, 3*1024*1024, n)
		if start.Len() < SEEK {
			start.Write(buff)
		}
		fmt.Println("Write", i)
		time.Sleep(time.Duration(r.Int63n(10)) * time.Millisecond)
	}
	assert.Equal(t, 30*1024*1024, l.Len())
	l.Close()

	h := sha256.New()
	n, err := h.Write(start.Bytes()[:SEEK]) // writing the first bite
	assert.NoError(t, err)
	assert.Equal(t, n, SEEK)
	err = l.Copy(SEEK, h) // writing the last bite
	assert.NoError(t, err)
	assert.Equal(t, l.Hash(), h.Sum(nil))

}
