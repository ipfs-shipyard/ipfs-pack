
all: deps

install: deps
	go install 
	
deps: gx
	gx install

gx:
	go get github.com/whyrusleeping/gx
	go get github.com/whyrusleeping/gx-go
