
# Makefile for github

VERSION ?= 0.0.0
BINARY_NAME ?= github

build:

test:

archive:

release:

build-prerequisites:
	mkdir -p bin dist

release-prerequisites:

test-prerequisites:

install-tools:

### BUILD ###################################################################

build-github: build-prerequisites
	go build -ldflags "-X main.version=${VERSION} -X main.commit=$$(git rev-parse --short HEAD 2>/dev/null || echo \"none\")" -o bin/$(OUTPUT_DIR)$(BINARY_NAME) cli/main.go
build-github-linux_amd64: build-prerequisites
	$(MAKE) GOOS=linux GOARCH=amd64 OUTPUT_DIR=linux_amd64/ build
build-github-darwin_amd64: build-prerequisites
	$(MAKE) GOOS=darwin GOARCH=amd64 OUTPUT_DIR=darwin_amd64/ build
build-github-windows_amd64: build-prerequisites
	$(MAKE) GOOS=windows GOARCH=amd64 OUTPUT_DIR=windows_amd64/ build

build-linux_amd64: build-github-linux_amd64
build-darwin_amd64: build-github-darwin_amd64
build-windows_amd64: build-github-windows_amd64

build: build-github
build-all: build-linux_amd64 build-darwin_amd64 build-windows_amd64

### ARCHIVE #################################################################

archive-github-linux_amd64: build-github-linux_amd64
	tar czf dist/$(BINARY_NAME)-${VERSION}-linux_amd64.tar.gz -C bin/linux_amd64/ .
archive-github-darwin_amd64: build-github-darwin_amd64
	tar czf dist/$(BINARY_NAME)-${VERSION}-darwin_amd64.tar.gz -C bin/darwin_amd64/ .
archive-github-windows_amd64: build-github-windows_amd64
	tar czf dist/$(BINARY_NAME)-${VERSION}-windows_amd64.tar.gz -C bin/windows_amd64/ .

archive-linux_amd64: archive-github-linux_amd64
archive-darwin_amd64: archive-github-darwin_amd64
archive-windows_amd64: archive-github-windows_amd64

archive: archive-linux_amd64 archive-darwin_amd64 archive-windows_amd64

release: archive
	sha1sum dist/*.tar.gz > dist/$(BINARY_NAME)-${VERSION}.shasums

### TEST ####################################################################

test-github:
	ginkgo
test-github-watch:
	ginkgo watch
test: test-github
.PHONY: test-github
.PHONY: test

clean:
	rm -r bin/* dist/*

### DATABASE ################################################################

db-up:
	psql < db/up.sql

db-down:
	psql < db/down.sql

db-seed:
	psql < db/seed.sql

