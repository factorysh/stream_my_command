package stream

import (
	"bytes"
	"fmt"
	"os"
	"path"
)

type Bucket struct {
	n      int
	file   *os.File
	buffer *bytes.Buffer
	home   string
	size   int
}

func NewBucket(home string, size int) (*Bucket, error) {
	b := &Bucket{
		n:      0,
		buffer: bytes.NewBuffer(nil),
		home:   home,
		size:   size,
	}
	b.buffer.Grow(size)
	err := b.reset()
	return b, err
}

func BucketPath(home string, n int) string {
	return path.Join(home, fmt.Sprintf("bucket_%d", n))
}

func (b *Bucket) reset() error {
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

func (b *Bucket) Len() int {
	return b.Len()
}

func (b *Bucket) maxChunkSize() int {
	return b.size - b.buffer.Len()
}

func (b *Bucket) write(chunk []byte) (int, error) {
	// asert len(chunk) <= maxChinkSize
	n, err := b.file.Write(chunk)
	if err != nil {
		return n, err
	}
	return b.buffer.Write(chunk)
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func (b *Bucket) Write(bite []byte) (int, error) {
	start := 0
	lbite := len(bite)
	for {
		size := min(b.maxChunkSize(), lbite-start)
		n, err := b.write(bite[start : start+size])
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

func (b *Bucket) Close() error {
	err := b.file.Chmod(0400)
	if err != nil {
		return err
	}
	b.buffer = nil // free some RAM
	return b.file.Close()
}

func (b *Bucket) Cache() []byte {
	return b.buffer.Bytes()
}
