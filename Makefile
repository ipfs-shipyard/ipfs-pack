export GOPATH=`pwd`/vendor

all:
	true

build: deps
	go build

deps:
	mkdir -p vendor/src/github.com/ipfs/
	git clone https://github.com/ipfs/go-ipfs vendor/src/github.com/ipfs/go-ipfs
	cd vendor/src/github.com/ipfs/go-ipfs
	git checkout feat/filestore0
	make deps
