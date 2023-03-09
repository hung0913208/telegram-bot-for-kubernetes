######################################################################
# @author      : Hung Nguyen Xuan Pham (hung0913208@gmail.com)
# @file        : Makefile
# @created     : Wednesday Jan 25, 2023 17:12:42 +07
######################################################################

GO := go

all: build test

build:
	@-$(GO) mod tidy
	@-$(GO) build -v ./...

test:
	@-$(GO) mod tidy
	@-$(GO) test -v ./tests/...
