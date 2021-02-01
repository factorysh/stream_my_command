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
	h, err := CommandHandler(command.Environment, command.Command, command.Arguments...)
	if err != nil {
		return err
	}
	server.HandleFunc(uri, h)
	return nil
}

func CommandHandler(env map[string]string, command string, args ...string) (http.HandlerFunc, error) {
	pool := _command.NewPool()
	arguments, err := _command.NewArguments(args)
	if err != nil {
		return nil, err
	}
	buffers := make(map[string]*stream.LongBuffer)
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
		longBuffer, ok := buffers[k]
		if !ok {
			longBuffer, err = stream.NewLongBuffer(os.TempDir())
			if err != nil {
				fmt.Println("error", err)
				w.WriteHeader(400)
				return
			}
			buffers[k] = longBuffer
			lock.Unlock()
			reader = longBuffer.Reader(0)

			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()
			w.Header().Add("Stream-Status", "fresh")
			w.WriteHeader(200)
			go func() {
				pool.Command(ctx, longBuffer, env, command, zargs...)
				longBuffer.Close()
			}()
		} else {
			lock.Unlock()
			reader = longBuffer.Reader(0)
			w.Header().Add("Stream-Status", "refurbished")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			w.WriteHeader(200)
		}
		io.Copy(w, reader)

	}, nil
}
