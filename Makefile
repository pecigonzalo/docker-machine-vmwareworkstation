export GO15VENDOREXPERIMENT = 1

default: build

build: clean
	go build -i -o ./bin/docker-machine-driver-vmwareworkstation.exe ./cmd/

clean:
	$(RM) bin/*

.PHONY: clean build