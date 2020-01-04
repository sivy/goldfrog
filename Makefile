SERVER_OUT := goldfrogd
INDEXER_OUT := indexer
PKG := github.com/sivy/goldfrog

TAG := $(shell git describe --tags)
VERSION := $(shell git describe --tags --long --always)

PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)

all: run

server:
#	go build -i -v -o dist/${SERVER_OUT} -ldflags="-X main.version=${VERSION}" ${PKG}
	go build -v -o dist/${SERVER_OUT} -ldflags="-X main.version=${VERSION} -X main.tag=${TAG}" cmd/goldfrogd/main.go

indexer:
#	go build -i -v -o dist/${INDEXER_OUT} -ldflags="-X main.version=${VERSION}" ${PKG}
	go build -v -o dist/${INDEXER_OUT} -ldflags="-X main.version=${VERSION} -X main.tag=${TAG}" cmd/indexer/main.go

test:
	go test -short ${PKG_LIST}

run: server
	./${SERVER_OUT}

.PHONY: run server
