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
	api.Register(mux, api.Command{
		Slug:      "nmap",
		Command:   "nmap",
		Arguments:   []string{"-A", "-T4", "-oX", "-", "$1"},
		ContentType: "application/xml",
	})
	http.Handle("/", mux)
	log.Fatal(http.ListenAndServe(":5000", nil))
}
```

the command is exposed as GET `/api/v1/nmap/{domain}`

#### Demo time

```
make
./bin/stream
```

In a another terminal
```
$ curl -v http://localhost:5000/api/v1/nmap/toto.com
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 5000 (#0)
> GET /api/v1/nmap/toto.com HTTP/1.1
> Host: localhost:5000
> User-Agent: curl/7.54.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: text/plain
< Stream-Status: fresh
< Date: Mon, 01 Feb 2021 20:37:47 GMT
< Transfer-Encoding: chunked
<
```

The first call trigger the command, the `Stream-Status` is `fresh`.

In a third terminal
```
$ curl -v http://localhost:5000/api/v1/nmap/toto.com
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 5000 (#0)
> GET /api/v1/nmap/toto.com HTTP/1.1
> Host: localhost:5000
> User-Agent: curl/7.54.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Stream-Status: refurbished
< Date: Mon, 01 Feb 2021 20:37:50 GMT
< Transfer-Encoding: chunked
<
```

The header `Stream-Status` says that you've got a second hand answer, and it will follow the main STDOUT stream.

You can kill a running command
```
curl -v -X DELETE http://localhost:5000/api/v1/nmap/toto.com
```
