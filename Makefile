SERVER_OUT := goldfrogd
INDEXER_OUT := indexer
PERSISTOR_OUT := persister
PKG := github.com/sivy/goldfrog

VERSION := $(shell git describe --tags --long --always)

PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)

all: run

server:
#	go build -i -v -o ${SERVER_OUT} -ldflags="-X main.version=${VERSION}" ${PKG}
	go build -v -o ${SERVER_OUT} -ldflags="-X main.version=${VERSION}" cmd/goldfrogd/main.go

indexer:
#	go build -i -v -o ${INDEXER_OUT} -ldflags="-X main.version=${VERSION}" ${PKG}
	go build -v -o ${INDEXER_OUT} -ldflags="-X main.version=${VERSION}" cmd/indexer/main.go

persister:
#	go build -i -v -o ${INDEXER_OUT} -ldflags="-X main.version=${VERSION}" ${PKG}
	go build -v -o ${PERSISTOR_OUT} -ldflags="-X main.version=${VERSION}" cmd/persister/main.go

test:
	go test -short ${PKG_LIST}

run: server
	./${SERVER_OUT}

.PHONY: run server
