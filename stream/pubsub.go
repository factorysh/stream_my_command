package stream

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

type Pubsub struct {
	lock        sync.Mutex
	subscribers map[int]*Subscriber
	buffer      *os.File
	size        int
}

func NewPubsub(home string) (*Pubsub, error) {
	f, err := ioutil.TempFile(home, "stream-")
	if err != nil {
		return nil, err
	}
	return &Pubsub{
		buffer:      f,
		lock:        sync.Mutex{},
		subscribers: make(map[int]*Subscriber),
	}, nil
}

type Subscriber struct {
	poz     int64
	in      io.ReadSeeker
	out     io.Writer
	fresh   chan interface{}
	ctx     context.Context
	writing bool
}

func (s *Subscriber) pump() {
	if !s.writing {
		s.fresh <- new(interface{})
	}
}

func (s *Subscriber) loop() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.fresh:

		}
		s.writing = true
		_, err := s.in.Seek(s.poz, os.SEEK_SET)
		if err != nil {
			panic(err)
		}
		n, err := io.Copy(s.out, s.in)
		if err != nil {
			panic(err)
		}
		s.poz += n
		s.writing = false
	}
}

func (p *Pubsub) Close() error {
	return p.buffer.Close()
}

// Write is publish
func (p *Pubsub) Write(buff []byte) (int, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	n, err := p.buffer.Write(buff)
	if err != nil {
		return n, err
	}
	p.size += n
	for _, s := range p.subscribers {
		s.pump()
	}
	return n, nil
}

func (p *Pubsub) Subscribe(ctx context.Context, seek int64, out io.Writer) error {
	p.lock.Lock()
	in, err := os.OpenFile(p.buffer.Name(), os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	s := &Subscriber{
		poz:     seek,
		out:     out,
		in:      in,
		ctx:     ctx,
		fresh:   make(chan interface{}),
		writing: false,
	}
	p.subscribers[len(p.subscribers)] = s
	p.lock.Unlock()
	go s.loop()
	return nil
}
