BINDIR ?= $(HOME)/.local/bin

.PHONY: build install install-fonts clean

build:
	go build -o bin/claude-pet-hook ./cmd/hook
	go build -o bin/claude-pet-statusline ./cmd/statusline

install: build
	cp bin/claude-pet-hook bin/claude-pet-statusline $(BINDIR)/

FONTDIR ?= $(HOME)/.local/share/fonts

install-fonts:
	mkdir -p $(FONTDIR)
	curl -fsSL -o $(FONTDIR)/Twemoji.Mozilla.ttf \
		https://github.com/mozilla/twemoji-colr/releases/latest/download/Twemoji.Mozilla.ttf
	fc-cache -f $(FONTDIR)
	@echo "Twemoji Mozilla installed. Add to your Ghostty config:"
	@echo "  font-family-emoji = Twemoji Mozilla"

clean:
	rm -rf bin/
