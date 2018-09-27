GO = go
REV = $(shell git rev-parse --short HEAD)
ENV = $(shell git --abbrev-ref HEAD)
VER = $(shell git describe --abbrev=0 --tags)
GOOS ?= linux
GOARCH ?= amd64
PKG = $(shell go list ./... | head -1)
PKGNAME = $(shell $(GO) list ./... | head -1 | sed -e 's/.*\///')

info:
	ECHO Rev: $(REV) Env: $(ENV) Ver: $(VER)
	ECHO $(GOOS) $(GOARCH) $(PKG) $(PKGNAME)

build:
	ECHO not implemented yet
