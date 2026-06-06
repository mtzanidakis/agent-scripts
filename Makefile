BIN_DIR := bin
BINARIES := gh-daily-tasks morning-tasks
PKG := ./...
LDFLAGS := -s -w

.PHONY: all build test vet fmt tidy lint clean $(BINARIES)

all: build

build: $(BINARIES)

$(BINARIES):
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$@ ./$@

test:
	go test $(PKG)

vet:
	go vet $(PKG)

fmt:
	gofmt -w .

tidy:
	go mod tidy

lint:
	golangci-lint run

clean:
	rm -rf $(BIN_DIR) dist
