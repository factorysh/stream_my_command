package main

import (
	"log"
	"net/http"

	"github.com/factorysh/stream_my_command/api"
)

func main() {
	mux := http.NewServeMux()
	api.Register(mux, api.Command{
		Slug:        "nmap",
		Command:     "nmap",
		Arguments:   []string{"-A", "-T4", "-oX", "-", "$1"},
		ContentType: "application/xml",
	})
	http.Handle("/", mux)
	log.Fatal(http.ListenAndServe(":5000", nil))
}
