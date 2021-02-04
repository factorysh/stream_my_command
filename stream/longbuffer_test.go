package stream

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	l, err := NewLongBuffer(os.TempDir())
	assert.NoError(t, err)

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
	l.Close()
	reader := l.Reader(0)
	b := new(bytes.Buffer)
	n, err := io.Copy(b, reader)
	assert.NoError(t, err)
	fmt.Println("path", l.path)
	assert.Equal(t, int64(30*1024*1024+512), n)
	h2 := sha256.New()
	h2.Write(b.Bytes())
	assert.Equal(t, h1.Sum(nil), l.Hash())
}

func TestLongBuffer(t *testing.T) {
	l, err := NewLongBuffer(os.TempDir())
	assert.NoError(t, err)

	r := rand.New(rand.NewSource(time.Now().Unix()))
	wg := sync.WaitGroup{}
	wg.Add(2)

	h3 := sha256.New()
	go func() { // Slow reader
		reader := l.Reader(0)
		b := new(bytes.Buffer)
		n := 0
		chunk := make([]byte, 512*1024)
		for {
			i, err := reader.Read(chunk)
			n += i
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
			b.Write(chunk[:i])
			time.Sleep(time.Duration(r.Int63n(30)) * time.Millisecond)
		}
		assert.NoError(t, err)
		assert.Equal(t, 30*1024*1024, n)
		h3.Write(b.Bytes())
		wg.Done()
	}()

	h4 := sha256.New()
	go func() {
		reader := l.Reader(0)
		b := new(bytes.Buffer)
		n, err := io.Copy(b, reader)
		assert.NoError(t, err)
		assert.Equal(t, int64(30*1024*1024), n)
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

	reader := l.Reader(0)
	b := new(bytes.Buffer)
	n, err := io.Copy(b, reader)
	assert.NoError(t, err)
	fmt.Println("path", l.path)
	assert.Equal(t, int64(30*1024*1024), n)
	h2 := sha256.New()
	h2.Write(b.Bytes())
	assert.Equal(t, l.Hash(), h1.Sum(nil))
	assert.Equal(t, l.Hash(), h2.Sum(nil))
	assert.Equal(t, l.Hash(), h3.Sum(nil))
	assert.Equal(t, l.Hash(), h4.Sum(nil))
}

func TestSeek(t *testing.T) {
	l, err := NewLongBuffer(os.TempDir())
	l.size = 1024
	assert.NoError(t, err)
	r := rand.New(rand.NewSource(time.Now().Unix()))
	buff := make([]byte, 3*1024)
	start := new(bytes.Buffer)
	SEEK := 5 * 1024
	for i := 0; i < 10; i++ {
		_, _ = r.Read(buff)
		n, err := l.Write(buff)
		assert.NoError(t, err)
		assert.Equal(t, 3*1024, n)
		if start.Len() < SEEK {
			start.Write(buff)
		}
		fmt.Println("Write", i)
		time.Sleep(time.Duration(r.Int63n(10)) * time.Millisecond)
	}
	assert.Equal(t, 30*1024, l.Len())
	l.Close()
	reader := l.Reader(SEEK)
	defer reader.Close()

	h := sha256.New()
	n, err := h.Write(start.Bytes()[:SEEK])
	assert.NoError(t, err)
	assert.Equal(t, n, SEEK)
	w, err := io.Copy(h, reader)
	assert.NoError(t, err)
	assert.Equal(t, w, int64(l.Len()-SEEK))
	assert.Equal(t, l.Hash(), h.Sum(nil))

}
