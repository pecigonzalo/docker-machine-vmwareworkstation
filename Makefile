#export GO15VENDOREXPERIMENT = 1

default: test build

test: clean
	go get -d -v -t ./...
	go test -v ./...
	go vet ./...

build: clean
	go get -d -v -t ./...
	go build -i -o ./bin/docker-machine-driver-vmwareworkstation.exe ./cmd/

clean:
	$(RM) bin/*

.PHONY: clean test build