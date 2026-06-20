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
