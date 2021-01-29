package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

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

func (s *Server) Register(name, command string, args ...string) error {
	uri := fmt.Sprintf("/api/v1/%s/", name)
	fmt.Println("uri", uri)
	s.routes.HandleFunc(uri, CommandHandler(command, args...))
	return nil
}

type WC struct {
	w io.Writer
}

func (wc *WC) Write(a []byte) (int, error) {
	return wc.w.Write(a)
}

func (wc *WC) Close() error {
	// nope
	return nil
}

func CommandHandler(command string, args ...string) http.HandlerFunc {
	pool := stream.New()
	return func(w http.ResponseWriter, r *http.Request) {
		slugs := strings.Split(r.URL.Path, "/")
		zargs := make([]string, len(args))
		for i, arg := range args {
			if arg[0] == '$' {
				n, err := strconv.Atoi(strings.TrimPrefix(arg, "$"))
				if err != nil {
					w.WriteHeader(500)
					return
				}
				zargs[i] = slugs[3+n]
			} else {
				zargs[i] = arg
			}
		}
		fmt.Println("zargs", zargs)

		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		w.WriteHeader(200)
		pool.Command(ctx, &WC{w}, command, zargs...)
	}
}
