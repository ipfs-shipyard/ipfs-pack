export GOPATH=$(shell pwd)/vendor

all:
	true

install: deps
	go install

build: deps
	go build

deps: vendor/src/github.com/ipfs/go-ipfs
	cd vendor/src/github.com/ipfs/go-ipfs && git checkout feat/filestore0 && make deps
	go get -d .

vendor/src/github.com/ipfs/go-ipfs:
	mkdir -p vendor
	ln -s $(GOPATH) vendor/src
	mkdir -p vendor/github.com/ipfs/
	git clone https://github.com/ipfs/go-ipfs vendor/github.com/ipfs/go-ipfs
