SRC := $(shell find . -name '*.go')
BINARY := docker-machine-driver-vmwareworkstation
LDFLAGS=-s -w

.PHONY: build
build: build/$(BINARY)-windows-amd64.exe

.PHONY: test
test:
	GOOS=windows GOARCH=amd64 go vet ./...
	GOOS=windows GOARCH=amd64 go test -v ./...

build/$(BINARY)-windows-amd64.exe: $(SRC)
	mkdir -p build
	GOOS=windows GOARCH=amd64 go build -o build/$(BINARY)-windows-amd64.exe ./cmd/plugin

.PHONY: clean
clean:
	-rm -rf build/*
