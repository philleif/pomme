.PHONY: build install clean run daemon

BINARY=pomme
BUILD_DIR=./build

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/pomme

install: build
	cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -rf $(BUILD_DIR)
	rm -f ~/.pomme/pomme.sock

run: build
	$(BUILD_DIR)/$(BINARY)

daemon: build
	$(BUILD_DIR)/$(BINARY) --daemon

deps:
	go mod tidy
	go mod download

test:
	go test ./...
