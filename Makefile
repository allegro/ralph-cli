# This Makefile is meant only for cross-compilation scenario, where we want to get
# binaries for all supported platforms at once.
# For other cases, use standard Go tooling (i.e., go build, go install).

PACKAGE_NAME := github.com/allegro/ralph-cli

deps:
	glide install

build-all: deps
	@echo "Building ralph-cli binaries for Darwin/Linux/Windows (64-bit)..."
	env GOOS=darwin GOARCH=amd64 go build $(PACKAGE_NAME) -o dist/ralph-cli-Darwin-x86_64
	env GOOS=linux GOARCH=amd64 go build $(PACKAGE_NAME) -o dist/ralph-cli-Linux-x86_64
	env GOOS=windows GOARCH=amd64 go build $(PACKAGE_NAME) -o dist/ralph-cli.exe

clean:
	rm -rf dist