package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	_command "github.com/factorysh/stream_my_command/command"
	"github.com/factorysh/stream_my_command/stream"
)

type Server struct {
	routes *http.ServeMux
}

func New() *Server {
	routes := http.NewServeMux()
	return &Server{
		routes: routes,
	}
}

func (s *Server) Mux() *http.ServeMux {
	return s.routes
}

func Register(server *http.ServeMux, name, command string, args ...string) error {
	uri := fmt.Sprintf("/api/v1/%s/", name)
	fmt.Println("uri", uri)
	h, err := CommandHandler(command, args...)
	if err != nil {
		return err
	}
	server.HandleFunc(uri, h)
	return nil
}

func CommandHandler(command string, args ...string) (http.HandlerFunc, error) {
	pool := _command.NewPool()
	arguments, err := _command.NewArguments(args)
	if err != nil {
		return nil, err
	}
	longBuffer, err := stream.NewLongBuffer(os.TempDir())
	if err != nil {
		return nil, err
	}

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

		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		w.WriteHeader(200)
		reader := longBuffer.Reader(0)
		go func() {
			pool.Command(ctx, longBuffer, nil, command, zargs...)
			longBuffer.Close()
		}()
		io.Copy(w, reader)

	}, nil
}
