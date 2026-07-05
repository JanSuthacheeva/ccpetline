VERSION ?= 0.0.7
BINDIR ?= $(HOME)/.local/bin
MODULE = github.com/jansuthacheeva/ccpetline
LDFLAGS = -ldflags "-X $(MODULE)/internal/pet.Version=$(VERSION)"
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

CMDS = hook statusline config
BINS = ccpetline-hook ccpetline ccpetline-config

.PHONY: build install clean release test lint

test:
	go test ./...

lint:
	go vet ./...
	@fmt=$$(gofmt -l .); if [ -n "$$fmt" ]; then \
		echo "gofmt needed on:"; echo "$$fmt"; exit 1; \
	fi

build:
	go build $(LDFLAGS) -o bin/ccpetline-hook ./cmd/hook
	go build $(LDFLAGS) -o bin/ccpetline ./cmd/statusline
	go build $(LDFLAGS) -o bin/ccpetline-config ./cmd/config

install: build
	mkdir -p $(BINDIR)
	cp bin/ccpetline-hook bin/ccpetline bin/ccpetline-config $(BINDIR)/

clean:
	rm -rf bin/ dist/

release: clean
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform#*/}; \
		echo "Building $$OS/$$ARCH..."; \
		EXT=""; \
		if [ "$$OS" = "windows" ]; then EXT=".exe"; fi; \
		GOOS=$$OS GOARCH=$$ARCH go build $(LDFLAGS) -o "dist/ccpetline-hook$${EXT}" ./cmd/hook && \
		GOOS=$$OS GOARCH=$$ARCH go build $(LDFLAGS) -o "dist/ccpetline$${EXT}" ./cmd/statusline && \
		GOOS=$$OS GOARCH=$$ARCH go build $(LDFLAGS) -o "dist/ccpetline-config$${EXT}" ./cmd/config && \
		if [ "$$OS" = "windows" ]; then \
			(cd dist && zip "ccpetline-$$OS-$$ARCH.zip" ccpetline-hook.exe ccpetline.exe ccpetline-config.exe && rm -f ccpetline-hook.exe ccpetline.exe ccpetline-config.exe); \
		else \
			(cd dist && tar czf "ccpetline-$$OS-$$ARCH.tar.gz" ccpetline-hook ccpetline ccpetline-config && rm -f ccpetline-hook ccpetline ccpetline-config); \
		fi; \
	done
	@echo "Release artifacts in dist/"
