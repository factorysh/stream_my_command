package stream

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"
)

type Bucket struct {
	n      int
	file   *os.File
	buffer *bytes.Buffer
	home   string
	size   int
	closed bool
	lock   *sync.RWMutex
}

func NewBucket(home string, size int) (*Bucket, error) {
	b := &Bucket{
		n:      0,
		buffer: bytes.NewBuffer(nil),
		home:   home,
		size:   size,
		lock:   &sync.RWMutex{},
	}
	b.buffer.Grow(size)
	err := b.reset()
	return b, err
}

func BucketPath(home string, n int) string {
	return path.Join(home, fmt.Sprintf("bucket_%d", n))
}

func (b *Bucket) reset() error {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.n++
	var err error
	if b.file != nil {
		err := b.file.Chmod(0400)
		if err != nil {
			return err
		}
		err = b.file.Close()
		if err != nil {
			return err
		}
	}
	b.file, err = os.OpenFile(BucketPath(b.home, b.n),
		os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	b.buffer.Reset()
	return nil
}

// Len length of all buckets
func (b *Bucket) Len() int {
	b.lock.RLock()
	defer b.lock.RUnlock()
	if b.closed {
		s, err := os.Stat(BucketPath(b.home, b.n))
		if err != nil {
			panic(err)
		}
		return ((b.n - 1) * b.size) + int(s.Size())
	}
	return ((b.n - 1) * b.size) + b.buffer.Len()
}

func (b *Bucket) lastBucketLen() int {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.buffer.Len()
}

func (b *Bucket) maxChunkSize() int {
	return b.size - b.buffer.Len()
}

func (b *Bucket) write(chunk []byte) (int, error) {
	// assert len(chunk) <= maxChinkSize
	return io.MultiWriter(b.file, b.buffer).Write(chunk)
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func (b *Bucket) Write(bite []byte) (int, error) {
	if b.closed {
		return 0, errors.New("Closed bucket")
	}
	start := 0
	lbite := len(bite)
	for {
		b.lock.Lock()
		size := min(b.maxChunkSize(), lbite-start)
		n, err := b.write(bite[start : start+size])
		b.lock.Unlock()
		if err != nil {
			return n, err
		}
		start += size
		if start == lbite {
			break
		}
		b.reset()
	}
	return start, nil
}

// Close the bucket
func (b *Bucket) Close() error {
	b.lock.Lock()
	defer b.lock.Unlock()
	err := b.file.Chmod(0400)
	if err != nil {
		return err
	}
	b.buffer = nil // free some RAM
	b.closed = true
	return b.file.Close()
}

// Cache return a copy of the last bucket buffer
func (b *Bucket) Cache() []byte {
	b.lock.RLock()
	defer b.lock.RUnlock()
	bb := b.buffer.Bytes()
	out := make([]byte, len(bb))
	copy(out, bb)
	return out
}

func (b *Bucket) seekMyCopy(seek int, w io.Writer) (int, error) {
	bucket := div(seek, b.size) + 1
	if bucket > b.n {
		if b.closed {
			return 0, fmt.Errorf("Bucket overflow %d/%d", bucket, b.n)
		}
		return 0, nil
	}
	start := seek - ((bucket - 1) * b.size)
	if !b.closed && bucket == b.n {
		if start > b.buffer.Len() {
			return 0, nil
		}
		return w.Write(b.Cache()[start:])
	}
	path := BucketPath(b.home, bucket)
	f, err := os.OpenFile(path, os.O_RDONLY, 0400)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		return 0, err
	}
	_, err = f.Seek(int64(start), io.SeekStart)
	if err != nil {
		return 0, err
	}
	if bucket == b.n && int(s.Size()) == start {
		return 0, io.EOF
	}
	size := 0
	for {
		p := make([]byte, int(s.Size())-start-size)
		n, err := f.Read(p)
		if err != nil {
			return size, err
		}
		_, err = w.Write(p)
		if err != nil {
			return size, err
		}
		size += n
		if int(s.Size()) == start+size {
			return size, nil
		}
	}
}

// Copy content of the bucket to a writer, waiting for fresh data
func (b *Bucket) Copy(start int, w io.Writer) error {
	for {
		n, err := b.seekMyCopy(start, w)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if n == 0 {
			fmt.Println("Waiting for data")
			time.Sleep(50 * time.Millisecond)
		}
		start += n
	}
}
