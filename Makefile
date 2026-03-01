BINDIR ?= $(HOME)/.local/bin

.PHONY: build install clean

build:
	go build -o bin/claude-pet-hook ./cmd/hook
	go build -o bin/claude-pet-statusline ./cmd/statusline
	go build -o bin/claude-pet-config ./cmd/config

install: build
	cp bin/claude-pet-hook bin/claude-pet-statusline bin/claude-pet-config $(BINDIR)/


clean:
	rm -rf bin/
