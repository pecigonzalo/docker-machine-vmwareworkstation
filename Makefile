BINARY := docker-machine-driver-vmwareworkstation
GOLANGCI_LINT := bin/tools/golangci-lint

# The driver only builds for Windows (registry and vmrun.exe access). The
# target is set per recipe so tool builds still use the host platform.
vet test build lint: export GOOS := windows
vet test build lint: export GOARCH := amd64

.PHONY: default
default: test build

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test: vet
ifeq ($(OS),Windows_NT)
	go test ./...
else
	@# Windows test binaries cannot run here; compile every package without running.
	go test -exec true ./...
endif

# golangci-lint is pinned in tools/go.mod and built for the host platform.
$(GOLANGCI_LINT): tools/go.mod tools/go.sum
	env -u GOOS -u GOARCH go build -modfile=tools/go.mod -o $@ github.com/golangci/golangci-lint/v2/cmd/golangci-lint

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

# Build flags live in .goreleaser.yaml so local, CI and release builds match.
.PHONY: build
build:
	goreleaser build --snapshot --clean --single-target -o bin/$(BINARY).exe

.PHONY: clean
clean:
	$(RM) -r bin dist
