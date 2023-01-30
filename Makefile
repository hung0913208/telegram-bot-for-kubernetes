######################################################################
# @author      : Hung Nguyen Xuan Pham (hung0913208@gmail.com)
# @file        : Makefile
# @created     : Wednesday Jan 25, 2023 17:12:42 +07
######################################################################

all: build test

build:
	go mod tidy
	go build -v ./...

test:
	go mod tidy
	go test -v ./test/...
