#export GO15VENDOREXPERIMENT = 1

default: test build

deps:
	go get github.com/Masterminds/glide
	glide install

test: deps
	go test -v ./...
	go vet ./...

build: deps
	go build -i -o ./bin/docker-machine-driver-vmwareworkstation.exe ./cmd/

clean:
	$(RM) -rf vendor
	$(RM) bin/*

.PHONY: clean test build