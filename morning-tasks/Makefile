.PHONY: build test

build:
	CGO_ENABLED=0 go build -ldflags='-s -w' -o morning-tasks .

test:
	go test -v ./...
