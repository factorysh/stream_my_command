package stream

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	_path "path"
	"sync"

	"github.com/google/uuid"
)

// Bucket handle one writer, multiple slow reader
type Bucket struct {
	id     uuid.UUID
	n      int
	file   *os.File
	buffer *bytes.Buffer
	home   string
	size   int
	closed bool
	lock   *sync.RWMutex
	hash   hash.Hash
}

// NewBucket returns a new Bucket, with its home and size
func NewBucket(home string, size int) (*Bucket, error) {
	id := uuid.New()
	path := _path.Join(home, fmt.Sprintf("lb-%s", id.String()))
	err := os.Mkdir(path, 0700)
	if err != nil {
		return nil, err
	}
	b := &Bucket{
		id:     id,
		n:      0,
		buffer: bytes.NewBuffer(nil),
		home:   path,
		size:   size,
		lock:   &sync.RWMutex{},
		hash:   sha256.New(),
	}
	b.buffer.Grow(size)
	err = b.reset()
	return b, err
}

// ID is the UUID
func (b *Bucket) ID() uuid.UUID {
	return b.id
}

// Path says where the storage folder is
func (b *Bucket) Path() string {
	return b.home
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
	return io.MultiWriter(b.file, b.buffer, b.hash).Write(chunk)
}

// Hash is the SHA256 of written datas
func (b *Bucket) Hash() []byte {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.hash.Sum(nil)
}

// Write a bite
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
		err = b.reset()
		if err != nil {
			return n, err
		}
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

func (b *Bucket) Closed() bool {
	return b.closed
}

// Cache return a copy of the last bucket buffer
func (b *Bucket) Cache(start int) []byte {
	b.lock.RLock()
	defer b.lock.RUnlock()
	bb := b.buffer.Bytes()
	if start > len(bb) {
		return []byte{}
	}
	out := make([]byte, len(bb)-start)
	copy(out, bb[start:])
	return out
}
