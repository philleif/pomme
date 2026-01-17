.PHONY: build install clean run daemon app app-install

BINARY=pomme
BUILD_DIR=./build
APP_NAME=Pomme.app
APP_DIR=./macos/$(APP_NAME)
APP_RESOURCES=$(APP_DIR)/Contents/Resources

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

# Build macOS app bundle
app: build
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(APP_RESOURCES)
	@cp $(BUILD_DIR)/$(BINARY) $(APP_RESOURCES)/$(BINARY)
	@chmod +x $(APP_DIR)/Contents/MacOS/launcher
	@echo "App bundle ready at $(APP_DIR)"

# Install app to /Applications
app-install: app
	@echo "Installing $(APP_NAME) to /Applications..."
	@rm -rf /Applications/$(APP_NAME)
	@cp -R $(APP_DIR) /Applications/
	@echo "Installed to /Applications/$(APP_NAME)"
