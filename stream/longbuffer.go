package stream

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"math"
	"os"
	"path"
	_path "path"
	"sync"
	"time"

	"github.com/google/uuid"
)

type LongBuffer struct {
	buffer   *bytes.Buffer
	lock     *sync.RWMutex
	path     string
	size     int
	len      int
	bucket   *os.File
	n_bucket int
	closed   bool
	id       uuid.UUID
	hash     hash.Hash
}

func NewLongBuffer(home string) (*LongBuffer, error) {
	id := uuid.New()
	path := _path.Join(home, fmt.Sprintf("lb-%s", id.String()))
	err := os.Mkdir(path, 0700)
	if err != nil {
		return nil, err
	}
	l := &LongBuffer{
		buffer: &bytes.Buffer{},
		lock:   &sync.RWMutex{},
		path:   path,
		size:   10 * 1024 * 1024,
		id:     id,
		hash:   sha256.New(),
	}
	err = l.newBucket()
	return l, err
}

func (l *LongBuffer) ID() uuid.UUID {
	return l.id
}

func (l *LongBuffer) Closed() bool {
	return l.closed
}

func (l *LongBuffer) newBucket() error {
	if l.bucket != nil {
		err := l.bucket.Close()
		if err != nil {
			return err
		}
	}
	f, err := os.OpenFile(l.bucketPath(l.n_bucket),
		os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	l.bucket = f
	l.n_bucket++
	l.buffer.Reset()
	return nil
}

func (l *LongBuffer) bucketPath(n int) string {
	return path.Join(l.path, fmt.Sprintf("bucket_%d", n))
}

func (l *LongBuffer) write(chunk []byte) (int, error) {
	_, err := l.buffer.Write(chunk)
	if err != nil {
		return 0, err
	}
	l.hash.Write(chunk)
	l.len += len(chunk)
	return l.bucket.Write(chunk)
}

func (l *LongBuffer) Len() int {
	return l.len
}

func (l *LongBuffer) Close() error {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.closed = true
	return nil
}

func (l *LongBuffer) Hash() []byte {
	return l.hash.Sum(nil)
}

func (l *LongBuffer) Write(blob []byte) (int, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.closed {
		return 0, errors.New("Closed buffer")
	}
	size := 0
	for {
		cSize := l.size - l.buffer.Len()
		if len(blob) < cSize {
			n, err := l.write(blob)
			if err != nil {
				return 0, err
			}
			return size + n, nil
		}
		_, err := l.write(blob[:cSize])
		if err != nil {
			return 0, err
		}
		err = l.newBucket()
		if err != nil {
			return 0, err
		}
		blob = blob[cSize:]
		size += cSize
	}
	return 0, nil
}

type LongBufferReader struct {
	l    *LongBuffer
	seek int
}

func div(x, y int) int {
	if x < y { // Early optimization
		return 0
	}
	return int(math.Floor(float64(x) / float64(y)))
}

func (r *LongBufferReader) Read(p []byte) (n int, err error) {
	if r.seek > r.l.len {
		if r.l.closed {
			return 0, fmt.Errorf("outside %d %d", r.seek, r.l.len)
		}
		time.Sleep(100 * time.Millisecond)
		return 0, nil
	}
	if r.l.closed && r.seek == r.l.len {
		return 0, io.EOF
	}
	bucket := div(r.seek, r.l.size)
	fmt.Println("seek", r.seek, "bucket", bucket, "/", r.l.n_bucket, r.seek-bucket*r.l.size)
	if bucket+1 == r.l.n_bucket {
		n = copy(r.l.buffer.Bytes(), p)
		fmt.Println("from cache", n)
		r.seek += n
		return n, nil
	}
	f, err := os.Open(r.l.bucketPath(bucket))
	if err != nil {
		return n, err
	}
	defer f.Close()
	_, err = f.Seek(int64(r.seek-bucket*r.l.size), io.SeekStart)
	if err != nil {
		return n, err
	}
	// FIXME, open/seek/close for each step is violent, add some lazyness
	n, err = f.Read(p)
	if err == io.EOF {
		err = nil // And we don't care!
	}
	if n == 0 {
		fmt.Println("Waiting")
		time.Sleep(100 * time.Millisecond)
	}

	r.seek += n
	return n, err
}

func (r *LongBufferReader) Close() error {
	return nil
}

func (l *LongBuffer) Reader(seek int) io.ReadCloser {
	return &LongBufferReader{
		l:    l,
		seek: seek,
	}
}
