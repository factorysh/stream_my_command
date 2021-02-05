package stream

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"time"
)

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
	r.l.lock.RLock()
	defer r.l.lock.RUnlock()
	if r.seek == r.l.len {
		if r.l.closed {
			return 0, io.EOF
		}
		time.Sleep(100 * time.Millisecond)
		return 0, nil
	}
	if r.seek > r.l.len {
		if r.l.closed {
			return 0, fmt.Errorf("outside %d %d", r.seek, r.l.len)
		}
		time.Sleep(100 * time.Millisecond)
		return 0, nil
	}
	bucket := div(r.seek, r.l.size) + 1 // 1…n counter
	start := r.seek - ((bucket - 1) * r.l.size)
	fmt.Println("seek", r.seek, "bucket", bucket, "/", r.l.n_bucket,
		start, r.l.closed)
	if !r.l.closed && bucket == r.l.n_bucket && false {
		f, err := os.Open(BucketPath(r.l.path, bucket))
		if err != nil {
			return n, err
		}
		s, err := f.Stat()
		if err != nil {
			return n, err
		}
		fmt.Println("mode", s.Mode())
		end := start + len(p)
		if end > r.l.buffer.Len() {
			end = r.l.buffer.Len()
		}
		bof := make([]byte, end-start)
		_, err = f.Seek(int64(start), io.SeekStart)
		if err != nil {
			return n, err
		}
		n, err = f.Read(bof)
		if err != nil {
			return n, err
		}
		slice := r.l.buffer.Bytes()[start:end]
		fmt.Println("bof", bucket, len(slice) == len(bof), len(bof), len(slice), bytes.Compare(bof, slice))
		if bytes.Compare(bof, slice) != 0 {
			for i, l := range slice {
				if bof[i] != l {
					fmt.Println("[file vs buffer]", i, bof[i], l)
					panic("argh")
				}
			}
		}
		fmt.Println("Stat", s.Size(), r.l.buffer.Len()-start, len(p))
		fmt.Println("Diff", r.l.buffer.Len(), s.Size(), r.l.buffer.Len() == int(s.Size()))
		n = copy(slice, p)
		fmt.Println("from cache", n, "start", start)
		r.seek += n
		return n, nil
	}
	f, err := os.Open(BucketPath(r.l.path, bucket))
	if err != nil {
		return n, err
	}
	defer f.Close()
	_, err = f.Seek(int64(start), io.SeekStart)
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
