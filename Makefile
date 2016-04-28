#!/usr/bin/make -f

SHELL=/bin/bash

deps:
	godep restore

build: deps golang-crosscompile
	source golang-crosscompile/crosscompile.bash; \
	go-darwin-amd64 build -o dist/ralph-scan-Darwin-x86_64; \
	go-linux-386 build -o dist/ralph-scan-Linux-i386; \
	go-linux-amd64 build -o dist/ralph-scan-Linux-x86_64; \
	go-windows-386 build -o dist/ralph-scan.exe

golang-crosscompile:
	git clone https://github.com/davecheney/golang-crosscompile.git

install:
	deps
