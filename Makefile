.PHONY: build run clean dev

# Full build: hash assets → render static HTML → compile server binary
build:
	go run build.go content.go templates.go render.go
	go build -o pixabros .

# Build and run
run: build
	./pixabros

# Development (skip build, just run server — requires dist/ already built)
dev:
	go build -o pixabros . && ./pixabros

# Clean build artifacts
clean:
	rm -rf dist/ pixabros

# Build assets + HTML only (no binary)
assets:
	go run build.go content.go templates.go render.go

# Self-contained Linux binary with all assets embedded
bundle-linux: assets
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o pixabros-linux .
	@echo "Built: pixabros-linux (static, embedded assets)"
