package stream

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync"
)

type Pool struct {
	lock sync.RWMutex
}

func New() *Pool {
	return &Pool{
		lock: sync.RWMutex{},
	}
}

func (p *Pool) Command(ctx context.Context, out io.WriteCloser, name string, args ...string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}
