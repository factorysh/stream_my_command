build: bin
	go build -o bin/stream cmd/stream.go

bin:
	mkdir -p bin
