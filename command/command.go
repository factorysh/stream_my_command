package command

import (
	"context"
	"fmt"
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

func (p *Pool) Command(ctx context.Context, out io.WriteCloser, env map[string]string, name string, args ...string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	defer out.Close()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	envs := make([]string, 0)
	for k, v := range env {
		// FIXME assert k ~=[a-zA-Z_][a-zA-Z_0-9]+
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = envs
	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}
