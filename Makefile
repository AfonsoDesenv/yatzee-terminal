GO := $(HOME)/.local/go/bin/go

.PHONY: run serve test test-race vet build

run:
	$(GO) run ./cmd/yatzee

serve:
	$(GO) run ./cmd/yatzee serve

test:
	$(GO) test ./...

test-race:
	@command -v gcc >/dev/null || { echo "precisa de compilador C: sudo apt install build-essential"; exit 1; }
	CGO_ENABLED=1 $(GO) test -race ./...

vet:
	$(GO) vet ./...

build:
	$(GO) build -o bin/yatzee ./cmd/yatzee
