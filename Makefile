SRC := $(shell find . -name '*.go')
BINARY := docker-machine-driver-vmwareworkstation
LD_FLAGS=-s -w

.PHONY: build
build: build/$(BINARY)-windows-amd64.exe

.PHONY: test
test:
	GOOS=windows GOARCH=amd64 go vet ./...
  GOOS=windows GOARCH=amd64 go test -race ./...

build/$(BINARY)-windows-amd64.exe: $(SRC)
	mkdir -p build
	GOOS=windows GOARCH=amd64 \
	go build -o build/$(BINARY)-windows-amd64.exe ./cmd/

.PHONY: clean
clean:
	-rm -rf build/
