BIN := gh-daily-tasks

.PHONY: build test
build:
	CGO_ENABLED=0 go build -ldflags='-s -w' -o $(BIN) .

test:
	go test -v -count=1 ./...
