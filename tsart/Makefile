GO ?= go
BIN_DIR := bin
TARGET := $(BIN_DIR)/tsart
TSART_CONFIG_DIR := ~/.config/tsart/
TSART_BINARY := ~/.config/tsart/tsart

.PHONY: build install

# build creates a host-native tsart executable at $(TARGET).
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(TARGET) .

# install builds tsart and installs it for the Neovim host configuration.
install: build
	mkdir -p $(TSART_CONFIG_DIR)
	cp $(TARGET) $(TSART_CONFIG_DIR)
