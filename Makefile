BIN ?= dct
INSTALL_FLAGS ?=

.PHONY: all build install deps clean

all: build

build:
	@mkdir -p bin
	@echo "Building $(BIN)..."
	@go build -o bin/$(BIN) .

install:
	@bash scripts/install.sh $(INSTALL_FLAGS)

deps:
	@bash -c 'for c in tar zip bzip2 md5sum sha1sum sha256sum go; do \
		if ! command -v "$$c" >/dev/null 2>&1; then \
			echo "MISS: $$c"; \
		else \
			echo "OK  : $$c"; \
		fi; \
	done'

clean:
	@rm -rf bin

