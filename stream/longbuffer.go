package stream

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

type LongBuffer struct {
	buffer   *bytes.Buffer
	lock     *sync.RWMutex
	path     string
	size     int
	len      int
	bucket   *os.File
	n_bucket int
}

func NewLongBuffer(home string) (*LongBuffer, error) {
	path, err := ioutil.TempDir(home, "lb-")
	if err != nil {
		return nil, err
	}
	l := &LongBuffer{
		buffer: &bytes.Buffer{},
		lock:   &sync.RWMutex{},
		path:   path,
		size:   10 * 1024 * 1024,
	}
	err = l.newBucket()
	return l, err
}

func (l *LongBuffer) newBucket() error {
	if l.bucket != nil {
		err := l.bucket.Close()
		if err != nil {
			return err
		}
	}
	f, err := os.OpenFile(path.Join(l.path, fmt.Sprintf("bucket_%d", l.n_bucket)),
		os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	l.bucket = f
	l.n_bucket++
	l.buffer.Reset()
	return nil
}

func (l *LongBuffer) write(chunk []byte) (int, error) {
	_, err := l.buffer.Write(chunk)
	if err != nil {
		return 0, err
	}
	l.len += len(chunk)
	return l.bucket.Write(chunk)
}

func (l *LongBuffer) Len() int {
	return l.len
}

func (l *LongBuffer) Write(blob []byte) (int, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
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
