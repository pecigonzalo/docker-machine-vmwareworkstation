#export GO15VENDOREXPERIMENT = 1

default: test build

test: clean
	go get -d -v -t ./...
	go get golang.org/x/tools/cmd/vet
	go test -v ./...
	go vet ./...

build: clean
	go get -d -v -t ./...
	go build -i -o ./bin/docker-machine-driver-vmwareworkstation.exe ./cmd/

clean:
	$(RM) bin/*

.PHONY: clean test build