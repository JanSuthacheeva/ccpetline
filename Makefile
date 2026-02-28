BINDIR ?= $(HOME)/.local/bin

.PHONY: build install run clean

build:
	go build -o bin/claude-pet-hook ./cmd/hook
	go build -o bin/claude-pet-statusline ./cmd/statusline

install: build
	cp bin/claude-pet-hook bin/claude-pet-statusline $(BINDIR)/

clean:
	rm -rf bin/
