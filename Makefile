export IPFS_API=ipfs.io

all: deps

install: deps
	go install 

build: deps
	go build

deps: gx
	gx install

gx:
	go get github.com/whyrusleeping/gx
	go get github.com/whyrusleeping/gx-go

test:
	cd test/sharness && make

.PHONY: test
