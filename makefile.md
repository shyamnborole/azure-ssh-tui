# Installation Makefile

You can copy the `install` target below into a standalone `Makefile` or just use these commands manually. 

To install the application using this file (if you rename it to `Makefile`), run:
```bash
make install
```

### Makefile snippet
```makefile
.PHONY: install build

APP_NAME = azure-ssh-tui
INSTALL_BIN = az-ssh
INSTALL_DIR = /opt/homebrew/bin

build:
	go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

install: build
	@echo "Installing $(APP_NAME) to $(INSTALL_DIR)/$(INSTALL_BIN)..."
	cp bin/$(APP_NAME) $(INSTALL_DIR)/$(INSTALL_BIN)
	@echo "Installation complete!"
	@echo "=========================================================="
	@echo "⚠️  IMPORTANT: Please run 'rehash' in your terminal now ⚠️"
	@echo "   to ensure your shell recognizes the new command."
	@echo "=========================================================="
```
