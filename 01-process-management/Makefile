# Process Management Demo Makefile

.PHONY: build run clean help

# Default target
help:
	@echo "🚀 Process Management Demo"
	@echo "========================="
	@echo ""
	@echo "Available targets:"
	@echo "  build    - Compile the demo program"
	@echo "  run      - Build and run the demo"
	@echo "  clean    - Remove built binaries"
	@echo "  help     - Show this help message"
	@echo ""
	@echo "Usage examples:"
	@echo "  make run     # Build and run the interactive demo"
	@echo "  make build   # Just compile the binary"

# Build the demo
build:
	@echo "🔨 Building process management demo..."
	go build -o process-demo demo.go
	@echo "✅ Build complete: ./process-demo"

# Run the demo
run: build
	@echo "🚀 Starting process management demo..."
	@echo "⚠️  Note: Some features may require elevated privileges"
	./process-demo

# Clean built files
clean:
	@echo "🧹 Cleaning up..."
	rm -f process-demo
	@echo "✅ Cleanup complete"

# Development targets
dev-run: build
	@echo "🛠️  Running in development mode..."
	sudo ./process-demo
