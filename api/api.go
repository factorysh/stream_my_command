package api

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	_command "github.com/factorysh/stream_my_command/command"
	"github.com/factorysh/stream_my_command/stream"
)

type Command struct {
	Slug        string
	Command     string
	Arguments   []string
	ContentType string
	Environment map[string]string
}

func Register(server *http.ServeMux, command Command) error {
	if command.ContentType == "" {
		command.ContentType = "text/plain"
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
		seek := 0
		rangeRaw := r.Header.Get("range")
		if rangeRaw != "" {
			rr := strings.Split(rangeRaw, "=")
			if len(rr) != 2 || rr[0] != "bytes" {
				w.WriteHeader(400)
				return
			}
			startend := strings.Split(rr[1], "-")
			seek, err = strconv.Atoi(startend[0])
			if err != nil {
				fmt.Println("error", err)
				w.WriteHeader(400)
				return
			}
			// FIXME handle end range
		}
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
			reader = longBuffer.Reader(seek)
			var ctx context.Context
			ctx, run.Cancel = context.WithCancel(context.TODO())
			w.Header().Set("Stream-Status", "fresh")
			go func() {
				pool.Command(ctx, longBuffer, c.Environment, c.Command, zargs...)
				longBuffer.Close()
			}()
		} else {
			lock.Unlock()
			if r.Method == "DELETE" {
				run.Cancel()
				w.Header().Set("X-Id", run.LongBuffer.ID().String())
				w.WriteHeader(200)
				return
			}
			if run.LongBuffer.Closed() {
				w.Header().Set("Content-Length", fmt.Sprintf("%d", run.LongBuffer.Len()-seek))
				w.Header().Set("etag", hex.EncodeToString(run.LongBuffer.Hash()))
			}
			w.Header().Set("Stream-Status", "refurbished")
			reader = run.LongBuffer.Reader(seek)
		}
		w.Header().Set("Content-Type", c.ContentType)
		w.Header().Set("Accept-Range", "bytes")
		w.Header().Set("X-Id", run.LongBuffer.ID().String())
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		w.WriteHeader(200)
		io.Copy(w, reader)
	}, nil
}
