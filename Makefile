BINDIR ?= $(HOME)/.local/bin

.PHONY: build install clean

build:
	go build -o bin/ccpetline-hook ./cmd/hook
	go build -o bin/ccpetline ./cmd/statusline
	go build -o bin/ccpetline-config ./cmd/config

install: build
	cp bin/ccpetline-hook bin/ccpetline bin/ccpetline-config $(BINDIR)/


clean:
	rm -rf bin/
