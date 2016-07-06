#export GO15VENDOREXPERIMENT = 1

default: deps test build

deps: clean
	go get github.com/Masterminds/glide
	glide install

test:
	go test -v $(glide novendor)
	go vet $(glide novendor)

build:
	go build -i -o ./bin/docker-machine-driver-vmwareworkstation.exe ./cmd/

clean:
	$(RM) -rf vendor
	$(RM) bin/*

.PHONY: clean deps test build