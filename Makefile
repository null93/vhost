.DEFAULT_GOAL := build
.PHONY: pretty clean optimized

build:
	GOOS=linux go build -ldflags="-s -w" -trimpath -o bin/vhost cmd/vhost/main.go

clean:
	rm -f bin/*

pretty:
	command -v goimports || go install golang.org/x/tools/cmd/goimports@latest
	gofmt -w -s cmd internal sdk
	goimports -w cmd internal sdk

optimized: clean build
	upx --best --lzma bin/*