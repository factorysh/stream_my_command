package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	_command "github.com/factorysh/stream_my_command/command"
	"github.com/factorysh/stream_my_command/stream"
)

type Command struct {
	Slug        string
	Command     string
	Arguments   []string
	MimeType    string
	Environment map[string]string
}

func Register(server *http.ServeMux, command Command) error {
	if command.MimeType == "" {
		command.MimeType = "text/plain"
	}

	uri := fmt.Sprintf("/api/v1/%s/", command.Slug)
	fmt.Println("uri", uri)
	h, err := command.Handler()
	if err != nil {
		return err
	}
	server.HandleFunc(uri, h)
	return nil
}

type Run struct {
	LongBuffer *stream.LongBuffer
	Cancel     context.CancelFunc
}

func (c *Command) Handler() (http.HandlerFunc, error) {
	pool := _command.NewPool()
	arguments, err := _command.NewArguments(c.Arguments)
	if err != nil {
		return nil, err
	}
	buffers := make(map[string]*Run)
	lock := &sync.RWMutex{}

	return func(w http.ResponseWriter, r *http.Request) {
		slugs := strings.Split(r.URL.Path, "/")
		fmt.Println("slugs", slugs)
		zargs, err := arguments.Values(slugs[4:])
		if err != nil {
			fmt.Println("error", err)
			w.WriteHeader(400)
			return
		}
		fmt.Println("zargs", zargs)
		k := strings.Join(zargs, "/")
		var reader io.ReadCloser
		lock.Lock()
		run, ok := buffers[k]
		if !ok {
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				lock.Unlock()
				return
			}
			longBuffer, err := stream.NewLongBuffer(os.TempDir())
			if err != nil {
				fmt.Println("error", err)
				w.WriteHeader(500)
				return
			}
			run = &Run{
				LongBuffer: longBuffer,
			}
			buffers[k] = run
			lock.Unlock()
			reader = longBuffer.Reader(0)
			var ctx context.Context
			ctx, run.Cancel = context.WithCancel(context.TODO())
			w.Header().Add("Stream-Status", "fresh")
			w.WriteHeader(200)
			go func() {
				pool.Command(ctx, longBuffer, c.Environment, c.Command, zargs...)
				longBuffer.Close()
			}()
		} else {
			lock.Unlock()
			if r.Method == "DELETE" {
				run.Cancel()
				w.WriteHeader(200)
				return
			}
			reader = run.LongBuffer.Reader(0)
			w.Header().Add("Stream-Status", "refurbished")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			w.WriteHeader(200)
		}
		io.Copy(w, reader)

	}, nil
}
