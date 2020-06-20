PKG := github.com/sivy/goldfrog

SERVER_OUT := goldfrogd
INDEXER_OUT := indexer
PERSISTER_OUT := persister
WM_OUT := wm

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
	go build -v -o ${PERSISTER_OUT} -ldflags="-X main.version=${VERSION}" cmd/persister/main.go

wm:
#	go build -i -v -o ${INDEXER_OUT} -ldflags="-X main.version=${VERSION}" ${PKG}
	go build -v -o ${WM_OUT} -ldflags="-X main.version=${VERSION}" cmd/wm/main.go

test:
	go test -short ${PKG_LIST}

run: server
	./${SERVER_OUT}

.PHONY: run server
