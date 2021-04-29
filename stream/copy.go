package stream

import (
	"fmt"
	"io"
	"os"
	"time"
)

func (b *Bucket) seekMyCopy(seek int, w io.Writer) (int, error) {
	nBucket := b.n
	d := div(seek, b.size)
	rest := seek - (d * b.size)
	bucket := d + 1
	if d == nBucket && rest == 0 && b.closed { // nothing else to do
		return 0, io.EOF
	}
	if d >= nBucket {
		if b.closed {
			return 0, fmt.Errorf("Bucket overflow %d/%d : %d",
				bucket, nBucket, rest)
		}
		return 0, nil
	}
	start := seek - ((bucket - 1) * b.size)
	if !b.closed && bucket == nBucket {
		cached := b.Cache(start)
		if len(cached) == 0 {
			return 0, nil
		}
		return w.Write(cached)
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
	if bucket == nBucket && int(s.Size()) == start {
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

// Copy content of the bucket to a writer, and waits for fresh data
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
			time.Sleep(10 * time.Millisecond)
		}
		start += n
	}
}
