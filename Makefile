GO=go
DEP=dep
REV=$(shell git rev-parse --short HEAD)
ENV=$(shell git rev-parse --abbrev-ref HEAD)
VER=$(shell git describe --abbrev=0 --tags)
GOOS?=linux
GOARCH?=amd64
PKG=$(shell go list ./... | head -1)
PKGNAME=$(shell $(GO) list ./... | head -1 | sed -e 's/.*\///')

info:
	@printf "Rev  $(REV)\nEnv  $(ENV)\nVer  $(VER)\nOS   $(GOOS)\nARCH $(GOARCH)\nPKG  $(PKG)\nNAME $(PKGNAME)\n"

build: 
	$(GO) build -ldflags="-X $(PKG).Version=$(VER) -X $(PKG).Revision=$(REV) -X $(PKG).Env=$(ENV)" \
		-o ./bin/$(GOOS)-$(GOARCH)-$(PKGNAME) ./internal/cmd/

run: build
	./bin/$(GOOS)-$(GOARCH)-$(PKGNAME)

config:
	cp example.config.json config.json

init: config
	$(DEP) ensure
