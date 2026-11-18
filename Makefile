BINARY := docker-machine-driver-vmwareworkstation
# The driver only builds for Windows (registry and vmrun.exe access).
export GOOS := windows
export GOARCH := amd64

.PHONY: default
default: test build

.PHONY: test
test:
	go vet ./...
ifeq ($(OS),Windows_NT)
	go test ./...
else
	@# Windows test binaries cannot run here; compile every package without running.
	go test -exec true ./...
endif

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: build
build:
	go build -trimpath -ldflags "-s -w" -o bin/$(BINARY).exe ./cmd/

.PHONY: clean
clean:
	$(RM) -r bin
