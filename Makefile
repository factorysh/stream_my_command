build: bin
	go build -o bin/stream cmd/stream.go

test:
	go test -v -timeout 30s \
		github.com/factorysh/stream_my_command/rfc7233 \
		github.com/factorysh/stream_my_command/command \
		github.com/factorysh/stream_my_command/stream

bin:
	mkdir -p bin
