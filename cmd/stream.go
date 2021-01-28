package main

import (
	"log"
	"net/http"

	"github.com/factorysh/stream_my_command/api"
)

func main() {
	s := api.New()
	s.Register("nmap", "nmap", "$1")
	http.Handle("/", s.Mux())
	log.Fatal(http.ListenAndServe(":5000", nil))
}
