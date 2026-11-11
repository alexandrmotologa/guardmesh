.PHONY: all build test clean run build-web

BINARY_NAME=guardmesh

all: build

build-web:
	cd web && npm install && npm run build

build:
	go build -o $(BINARY_NAME) ./cmd/guardmesh

test:
	go test -v -race ./...

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -f coverage.out coverage.html
	rm -rf web/dist
