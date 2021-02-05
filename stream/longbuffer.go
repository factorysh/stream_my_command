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

type LongBuffer struct {
	buffer   *bytes.Buffer
	lock     *sync.RWMutex
	path     string
	size     int
	len      int
	bucket   *Bucket
	n_bucket int
	closed   bool
	id       uuid.UUID
	hash     hash.Hash
	writer   io.Writer
}

func NewLongBuffer(home string) (*LongBuffer, error) {
	id := uuid.New()
	path := _path.Join(home, fmt.Sprintf("lb-%s", id.String()))
	err := os.Mkdir(path, 0700)
	if err != nil {
		return nil, err
	}
	size := 10 * 1024 * 1024
	b, err := NewBucket(path, size)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	return &LongBuffer{
		path:   path,
		lock:   &sync.RWMutex{},
		bucket: b,
		id:     id,
		hash:   h,
		size:   size,
		writer: io.MultiWriter(b, h),
	}, err
}

func (l *LongBuffer) ID() uuid.UUID {
	return l.id
}

func (l *LongBuffer) Closed() bool {
	return l.closed
}

func (l *LongBuffer) Len() int {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.len
}

func (l *LongBuffer) Close() error {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.closed = true
	return l.bucket.Close()
}

func (l *LongBuffer) Hash() []byte {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.hash.Sum(nil)
}

func (l *LongBuffer) Write(blob []byte) (int, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.closed {
		return 0, errors.New("Closed buffer")
	}
	n, err := l.writer.Write(blob)
	l.len += n
	return n, err
}
