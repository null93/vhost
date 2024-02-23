.DEFAULT_GOAL := build
.PHONY: pretty clean build test package tools

tools:
	@echo "Installing tools..."
	command -v goreleaser || go install github.com/goreleaser/goreleaser@latest
	command -v goimports || go install golang.org/x/tools/cmd/goimports@latest

deps:
	@echo "Downloading dependencies..."
	go mod download

build:
	@echo "Building for your arch..."
	goreleaser build --snapshot --clean
	
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
	goreleaser release --clean
