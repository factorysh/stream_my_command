package api

import (
	"context"
	"encoding/hex"
	"fmt"
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
	Bucket *stream.Bucket
	Cancel context.CancelFunc
}

func simpleStartRange(rangeRaw string) (int, error) {
	rr := strings.Split(rangeRaw, "=")
	if len(rr) != 2 || rr[0] != "bytes" {
		return 0, fmt.Errorf("Not bytes : %s", rangeRaw)
	}
	startend := strings.Split(rr[1], "-")
	seek, err := strconv.Atoi(startend[0])
	if err != nil {
		fmt.Println("error", err)
		return 0, fmt.Errorf("Bof")
	}
	// FIXME handle end range
	return seek, nil
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
		seek := 0
		rangeRaw := r.Header.Get("range")
		if rangeRaw != "" {
			seek, err = simpleStartRange(rangeRaw)
			if err != nil {
				w.WriteHeader(400)
				return
			}
		}
		lock.Lock()
		run, ok := buffers[k]
		if !ok {
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				lock.Unlock()
				return
			}
			longBuffer, err := stream.NewBucket(os.TempDir(), 10*1024*1024)
			if err != nil {
				fmt.Println("error", err)
				w.WriteHeader(500)
				return
			}
			run = &Run{
				Bucket: longBuffer,
			}
			buffers[k] = run
			lock.Unlock()
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
				w.Header().Set("X-Id", run.Bucket.ID().String())
				w.WriteHeader(200)
				return
			}
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if run.Bucket.Closed() {
				w.Header().Set("Content-Length", fmt.Sprintf("%d", run.Bucket.Len()-seek))
				w.Header().Set("etag", hex.EncodeToString(run.Bucket.Hash()))
			}
			w.Header().Set("Stream-Status", "refurbished")
		}
		w.Header().Set("Content-Type", c.ContentType)
		w.Header().Set("Accept-Range", "bytes")
		w.Header().Set("X-Id", run.Bucket.ID().String())
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		run.Bucket.Copy(seek, w)
	}, nil
}
