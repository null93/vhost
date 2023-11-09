.DEFAULT_GOAL := build
.PHONY: pretty clean build test package tools

tools:
	@echo "Installing tools..."
	command -v nfpm || go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
	command -v goimports || go install golang.org/x/tools/cmd/goimports@latest

deps:
	@echo "Downloading dependencies..."
	go mod download

build:
	@echo "Building for your arch..."
	GOOS=linux go build -ldflags="-s -w" -trimpath -o bin/vhost cmd/vhost/main.go
	
test:
	@echo "Running tests..."
	go test -v ./pkg/...

clean:
	@echo "Cleaning up..."
	-rm -rf bin dist

pretty: tools
	@echo "Making it all pretty..."
	gofmt -w -s cmd internal pkg
	goimports -w cmd internal pkg

package: clean pretty build
	@echo "Packaging for your arch..."
	mkdir -p dist
	nfpm pkg --target dist/vhost_0.0.1_linux_arm64.deb
