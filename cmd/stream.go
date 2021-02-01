package main

import (
	"log"
	"net/http"

	"github.com/factorysh/stream_my_command/api"
)

func main() {
	mux := http.NewServeMux()
	api.Register(mux, "nmap", "nmap", "$1")
	http.Handle("/", mux)
	log.Fatal(http.ListenAndServe(":5000", nil))
}
