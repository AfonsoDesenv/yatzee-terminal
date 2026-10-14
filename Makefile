GO := $(HOME)/.local/go/bin/go

.PHONY: run serve test vet build

run:
	$(GO) run ./cmd/yatzee

serve:
	$(GO) run ./cmd/yatzee serve

test:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

build:
	$(GO) build -o bin/yatzee ./cmd/yatzee
