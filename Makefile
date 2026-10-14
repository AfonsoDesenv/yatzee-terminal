GO := $(HOME)/.local/go/bin/go

.PHONY: run test test-race vet build build-windows

run:
	$(GO) run ./cmd/yatzee

test:
	$(GO) test ./...

test-race:
	@command -v gcc >/dev/null || { echo "precisa de compilador C: sudo apt install build-essential"; exit 1; }
	CGO_ENABLED=1 $(GO) test -race ./...

vet:
	$(GO) vet ./...

build:
	$(GO) build -o bin/yatzee ./cmd/yatzee

build-windows:
	GOOS=windows GOARCH=amd64 $(GO) build -o bin/yatzee.exe ./cmd/yatzee
