SRC := $(shell find . -name '*.go')
BINARY := bk
LD_FLAGS=-s -w

.PHONY: build
build: build/$(BINARY)-windows-amd64.exe

build/$(BINARY)-windows-amd64.exe: $(SRC)
	mkdir -p build
	GOOS=windows GOARCH=amd64 \
	go build -o build/$(BINARY)-windows-amd64.exe -ldflags="$(LD_FLAGS)" ./cmd/

.PHONY: clean
clean:
	-rm -rf build/
