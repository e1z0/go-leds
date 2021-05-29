all: build

deps:
	go get github.com/fulr/spidev
	go get github.com/gorilla/mux
	go get github.com/grandcat/zeroconf
	go get github.com/op/go-logging
	go get golang.org/x/text/encoding
	go get golang.org/x/text/encoding/charmap
	go get golang.org/x/text/transform

build:
	go build -o go_ledai .

.PHONY: build deps
