Stream my command
=================

Do REST stuff with commands.

## Command

An UNIX command is exposed, with few arguments.

Command run is a singleton, and try to end, even if HTTP call is interrupted.

Parralel call, or reconnection connect to the current STDOUT flow.

### Demo command

```golang
func main() {
	mux := http.NewServeMux()
	api.Register(mux, "nmap", "nmap", "-A", "-T4", "$1")
	http.Handle("/", mux)
	log.Fatal(http.ListenAndServe(":5000", nil))
}
```

the command is exposed as GET `/api/v1/nmap/{domain}`
