MAKEFLAGS += --silent

ldflags := -X 'github.com/telmo/mangal/constant.BuiltAt=$(shell date -u)'
ldflags += -X 'github.com/telmo/mangal/constant.BuiltBy=$(shell whoami)'
ldflags += -X 'github.com/telmo/mangal/constant.Revision=$(shell git rev-parse --short HEAD)'
ldflags += -s
ldflags += -w

build_flags := -ldflags=${ldflags}

all: help

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build the mangal binary"
	@echo "  install      Install the mangal binary"
	@echo "  uninstall    Uninstall the mangal binary"
	@echo "  test         Run the tests"
	@echo "  gif          Generate usage gifs"
	@echo "  mangadex     Build the mangadex_search binary"
	@echo "  help         Show this help message"
	@echo ""

install:
	@go install "$(build_flags)"

build:
	@go build "$(build_flags)"

mangadex:
	@go build "$(build_flags)" -o mangadex_search ./cmd/mangadex_search

test:
	@go test ./...

uninstall:
	@rm -f $(shell which mangal)

gif:
	@vhs assets/tui.tape
	@vhs assets/inline.tape
